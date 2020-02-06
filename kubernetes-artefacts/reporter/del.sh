delete from cpu where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from disk where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from diskio where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from interrupts where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from java_class_loading where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from java_garbage_collector where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from java_last_garbage_collection where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from java_memory where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from java_memory_pool where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from java_runtime where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from java_threading where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from kernel where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from linux_sysctl_fs where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from mem where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from net where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from netstat where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from processes where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from soft_interrupts where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from swap where "host"='you-jmeter-slave-7579df9dcc-d4x74';
delete from system where "host"='you-jmeter-slave-7579df9dcc-d4x74';


SHOW TAG VALUES FROM system WITH KEY=host