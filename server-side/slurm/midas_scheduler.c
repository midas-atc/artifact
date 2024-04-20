/*

This scheduler automates the allocation of smaller jobs to a dedicated resource group,
removing the requirement for manual submission of resource partitions by users.

*/

#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "slurm/slurm.h"
#include "slurm/slurm_errno.h"
#include "src/slurmctld/burst_buffer.h"
#include "src/slurmctld/fed_mgr.h"
#include "src/slurmctld/job_scheduler.h"
#include "src/slurmctld/locks.h"
#include "src/slurmctld/node_scheduler.h"
#include "src/slurmctld/preempt.h"
#include "src/slurmctld/reservation.h"
#include "src/slurmctld/slurmctld.h"
#include "src/slurmctld/srun_comm.h"

/* Local Variables */
static bool stop_midas = false;
static pthread_mutex_t term_lock = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t term_cond = PTHREAD_COND_INITIALIZER;
static bool config_flag = false;
static int midas_interval = MIDAS_INTERVAL;
static char *midas_params = DEFAULT_MIDAS_PARAMS;
static int max_sched_job_cnt = 1;
static int sched_timeout = 0;
static int num_resource_groups = 3; // Default number of resource groups

/* Local Functions Declarations */
static void begin_scheduling(void);
static void load_config(void);
static void my_sleep(int secs);
static int resource_group_allocation(job_queue_rec_t *job_rec);

/* External Function to Terminate midas_agent */
extern void stop_midas_agent(void) {
    slurm_mutex_lock(&term_lock);
    stop_midas = true;
    slurm_cond_signal(&term_cond);
    slurm_mutex_unlock(&term_lock);
}

/* Helper Function for Sleep */
static void my_sleep(int secs) {
    struct timespec ts = {0, 0};
    struct timeval now;

    gettimeofday(&now, NULL);
    ts.tv_sec = now.tv_sec + secs;
    ts.tv_nsec = now.tv_usec * 1000;
    slurm_mutex_lock(&term_lock);
    if (!stop_midas)
        slurm_cond_timedwait(&term_cond, &term_lock, &ts);
    slurm_mutex_unlock(&term_lock);
}

/* Load MIDAS Configuration */
static void load_config(void) {
    char *sched_params, *select_type, *tmp_ptr;

    sched_timeout = slurm_get_msg_timeout() / 2;
    sched_timeout = MAX(sched_timeout, 1);
    sched_timeout = MIN(sched_timeout, 10);

    sched_params = slurm_get_sched_params();

    if (sched_params && (tmp_ptr = strstr(sched_params, "interval=")))
        midas_interval = atoi(tmp_ptr + 9);
    if (midas_interval < 1) {
        error("Invalid SchedulerParameters interval: %d", midas_interval);
        midas_interval = MIDAS_INTERVAL;
    }

    if (sched_params && (tmp_ptr = strstr(sched_params, "max_job_bf=")))
        max_sched_job_cnt = atoi(tmp_ptr + 11);
    if (sched_params && (tmp_ptr = strstr(sched_params, "bf_max_job_test=")))
        max_sched_job_cnt = atoi(tmp_ptr + 16);
    if (max_sched_job_cnt < 1) {
        error("Invalid SchedulerParameters bf_max_job_test: %d", max_sched_job_cnt);
        max_sched_job_cnt = 50;
    }
    xfree(sched_params);

    select_type = slurm_get_select_type();
    if (!xstrcmp(select_type, "select/serial")) {
        max_sched_job_cnt = 0;
        stop_midas_agent();
    }
    xfree(select_type);

    midas_params = slurm_get_midas_params();
    debug2("MIDAS: loaded MIDAS params: %s", midas_params);

    // Load the number of resource groups
    if (sched_params && (tmp_ptr = strstr(sched_params, "num_resource_groups=")))
        num_resource_groups = atoi(tmp_ptr + 20);
    if (num_resource_groups < 1) {
        error("Invalid SchedulerParameters num_resource_groups: %d", num_resource_groups);
        num_resource_groups = 3; // Default to 3 resource groups
    }
}

static int resource_group_allocation(job_queue_rec_t *job_rec) {
    struct job_record *job_ptr = job_rec->job_ptr;
    int min_nodes = job_ptr->details->min_nodes;
    int group_size = node_record_count / num_resource_groups;

    if (min_nodes <= group_size) {
        return 1;
    } else if (min_nodes <= 2 * group_size) {
        return 2; 
    }
}

/* Composite Comparator for Sorting Based on MIDAS Parameters */
static int composite_comparator(void *x, void *y) {
    job_queue_rec_t *job_rec1 = *(job_queue_rec_t **) x;
    job_queue_rec_t *job_rec2 = *(job_queue_rec_t **) y;

    int result = 0;

    // Compare based on resource group allocation first
    int group1 = resource_group_allocation(job_rec1);
    int group2 = resource_group_allocation(job_rec2);

    if (group1 != group2) {
        return group1 - group2;
    }

    char *temp_midas_params = xmalloc(sizeof(*midas_params));
    strcpy(temp_midas_params, midas_params);


    if (result == 0)
        result = submit_time_comparator(job_rec1, job_rec2);

    xfree(temp_midas_params);

    return result;
}

/* Scheduling Logic */
static void begin_scheduling(void) {
    int j, rc = SLURM_SUCCESS, job_cnt = 0;
    List job_queue;
    job_queue_rec_t *job_queue_rec;
    struct job_record *job_ptr;
    struct part_record *part_ptr;
    bitstr_t *alloc_bitmap = NULL, *avail_bitmap = NULL;
    bitstr_t *exc_core_bitmap = NULL;
    uint32_t max_nodes, min_nodes;
    time_t now = time(NULL), sched_start;
    bool resv_overlap = false;
    sched_start = now;
    alloc_bitmap = bit_alloc(node_record_count);
    job_queue = build_job_queue(true, false);

    /* Sort Job Queue Based on MIDAS Parameters */
    debug2("MIDAS: current MIDAS params: %s", midas_params);
    list_sort(job_queue, composite_comparator);

    while ((job_queue_rec = (job_queue_rec_t *) list_pop(job_queue))) {
        job_ptr = job_queue_rec->job_ptr;
        part_ptr = job_queue_rec->part_ptr;
        xfree(job_queue_rec);
        if (part_ptr != job_ptr->part_ptr)
            continue; 

        if (++job_cnt > max_sched_job_cnt) {
            debug2("scheduling loop exiting after %d jobs", max_sched_job_cnt);
            break;
        }

        min_nodes = MAX(job_ptr->details->min_nodes, part_ptr->min_nodes);

        if (job_ptr->details->max_nodes == 0)
            max_nodes = part_ptr->max_nodes;
        else
            max_nodes = MIN(job_ptr->details->max_nodes, part_ptr->max_nodes);

        if (min_nodes > max_nodes) {
            continue;
        }

        j = job_test_resv(job_ptr, &now, true, &avail_bitmap, &exc_core_bitmap, &resv_overlap, false);

        if (j != SLURM_SUCCESS) {
            FREE_NULL_BITMAP(avail_bitmap);
            FREE_NULL_BITMAP(exc_core_bitmap);
            continue;
        }

        rc = select_nodes(job_ptr, false, NULL, NULL, false);

        if (rc == SLURM_SUCCESS) {
            /* Job initiated */
            last_job_update = time(NULL);
            debug2("MIDAS: Started JobId %d on %s", job_ptr->job_id, job_ptr->nodes);
            if (job_ptr->batch_flag == 0)
                srun_allocate(job_ptr);
            else if (!IS_JOB_CONFIGURING(job_ptr))
                launch_job(job_ptr);
        }

        FREE_NULL_BITMAP(avail_bitmap);
        FREE_NULL_BITMAP(exc_core_bitmap);

        if ((time(NULL) - sched_start) >= sched_timeout) {
            debug2("scheduling loop exiting after %d jobs", max_sched_job_cnt);
            break;
        }
    }

    FREE_NULL_LIST(job_queue);
    FREE_NULL_BITMAP(alloc_bitmap);
}

/* Function to Handle slurm.conf Changes */
extern void midas_reconfig(void) {
    config_flag = true;
}

/* MIDAS Scheduler Agent - Detached Thread */
extern void *midas_agent(void *args) {
    time_t now;
    double wait_time;
    static time_t last_sched_time = 0;
    /* Read config, nodes, and partitions; Write jobs */
    slurmctld_lock_t all_locks = {
        READ_LOCK, WRITE_LOCK, READ_LOCK, READ_LOCK, READ_LOCK };

    load_config();
    last_sched_time = time(NULL);
    while (!stop_midas) {
        my_sleep(midas_interval);
        if (stop_midas)
            break;
        if (config_flag) {
            config_flag = false;
            load_config();
        }
        now = time(NULL);
        wait_time = difftime(now, last_sched_time);
        if ((wait_time < midas_interval))
            continue;

        lock_slurmctld(all_locks);
        begin_scheduling();
        last_sched_time = time(NULL);
        (void) bb_g_job_try_stage_in();
        unlock_slurmctld(all_locks);
    }
    return NULL;
}
