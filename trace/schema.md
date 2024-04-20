Note: the actual trace data is pending approval from the university and will be released once approved.

### Job Submission History

| Timestamp (Job Submission) | Timestamp (Start) | Timestamp (Completion) | Resource Requests         | Node ID | Hashed User Name | Hashed Job Name | Software Dependencies |
|----------------------------|-------------------|------------------------|---------------------------|---------|------------------|-----------------|----------------------|
| 2022-01-01T09:00:00Z       | 2022-01-01T09:05:00Z | 2022-01-01T09:30:00Z | GPUs: 2, Memory: 16GB     | 1234    | ABCD1234         | EFGH5678        | TensorFlow, CUDA     |
| 2022-02-03T14:20:00Z       | 2022-02-03T14:25:00Z | 2022-02-03T15:00:00Z | GPUs: 1, Memory: 8GB      | 5678    | DCBA4321         | GFED8765        | PyTorch, CUDA        |
| 2022-03-15T11:45:00Z       | 2022-03-15T11:50:00Z | 2022-03-15T12:15:00Z | GPUs: 4, Memory: 32GB     | 91011   | QWERT0987        | UIOP7654        | Keras, OpenCL        |

### Node Status (aggregated every 5 minutes)

| Node ID | Time Stamp          | Duration (5 min) | Avg CPU Utilization | Max CPU Utilization | Avg Memory Utilization | Max Memory Utilization | Avg GPU Utilization | Max GPU Utilization | Inbound Traffic (Bytes) | Outbound Traffic (Bytes) | Networked File Storage IO (Bytes) |
|---------|---------------------|------------------|---------------------|---------------------|------------------------|------------------------|---------------------|---------------------|------------------------|-------------------------|----------------------------------|
| 1234    | 2022-01-01T09:00:00Z | 5 mins           | 60%                 | 80%                 | 70%                    | 90%                    | 50%                 | 70%                 | 100000000              | 80000000                | 50000000                         |
| 5678    | 2022-02-03T14:20:00Z | 5 mins           | 50%                 | 70%                 | 60%                    | 80%                    | 40%                 | 60%                 | 80000000               | 60000000                | 40000000                         |
| 91011   | 2022-03-15T11:45:00Z | 5 mins           | 70%                 | 90%                 | 80%                    | 95%                    | 60%                 | 80%                 | 120000000              | 100000000               | 60000000                         |

### Node Specifications

| Node ID | GPU Model | Number of GPUs | Total Memory |
|---------|-----------|----------------|--------------|
| 1234    | GTX 1080  | 2              | 16GB         |
| 5678    | RTX 3090  | 1              | 8GB          |
| 91011   | Tesla V100| 4              | 32GB         |

