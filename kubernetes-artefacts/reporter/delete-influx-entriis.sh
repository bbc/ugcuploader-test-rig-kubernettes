items_to_delete=(simon-jmeter-slave-6d79d6c4b9-gwss7,simon-jmeter-slave-6d79d6c4b9-rhw52,simon-jmeter-slave-6d79d6c4b9-w29qd,simon-jmeter-slave-6d79d6c4b9-xlmtt)
for i in ${items_to_delete[@]}; do
    influx -execute "delete from cpu where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from disk where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from diskio where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from interrupts where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from java_class_loading where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from java_garbage_collector where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from java_last_garbage_collection where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from java_memory where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from java_memory_pool where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from java_runtime where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from java_threading where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from kernel where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from linux_sysctl_fs where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from mem where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from net where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from netstat where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from processes where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from soft_interrupts where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from swap where host='$i'" -database jmeter-slaves
    echo "done"
    influx -execute "delete from system where host='$i'" -database jmeter-slaves
    echo "done"

done
