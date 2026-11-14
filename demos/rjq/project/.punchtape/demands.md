rjq speaks from a command line: rjq-cli <redis-url> <queue> <uuid> asks the queue for a job's status;
a failed ask is loud: the exact error line on stdout, and the exit code is one.
