redef Intel::item_expiration = 1day;

@if ( ! Cluster::is_enabled() 
    || Cluster::local_node_type() == Cluster::MANAGER )
event bro_init()
    {
    # Schedule the Abuse.ch ransomware intel sync
    local c = cron::CronJob($command=Exec::Command($cmd="./ransomware", $i=1hr, $reschedule=T));
    event cron::run_cron(c);
    }
@endif
