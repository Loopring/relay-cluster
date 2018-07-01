#!/bin/bash
#ValidateService

WORK_DIR=/opt/loopring/relay

#cron and logrotate are installed by default in ubuntu, don't check it again
if [ ! -f /etc/logrotate.d/loopring-relay ]; then
    sudo cp $WORK_DIR/src/bin/logrotate/loopring-relay /etc/logrotate.d/loopring-relay
fi

pgrep cron
if [[ $? != 0 ]]; then
    sudo /etc/init.d/cron start
fi
