module brointelutils;

export {
    ## Options for OTX sync
    const enable_otx: bool = F &redef;
    const otx_api_key: string = "" &redef;
    const otx_days: count 30 &redef;
    const otx_doNotice: bool = T &redef;
}

@if ( ! Cluster::is_enabled() 
    || Cluster::local_node_type() == Cluster::MANAGER ) 
event bro_init()
    {
    # Schedule OTX for 
    local otxCmd = fmt("%s/otx -apiKey %s -days %s", @DIR, otx_api_key, otx_days);
    if ( otx_doNotice )
        otxCmd = otxCmd + " -doNotice";
    
    if ( enable_otx )
        {
        local c = cron::CronJob(
            $command=Exec::Command($cmd=otxCmd), 
            $i=1hr, 
            $reschedule=T);
        event cron::run_cron(c);
        }
    }
@endif
