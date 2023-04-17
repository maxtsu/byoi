pid=`ps -ef | grep ".*app" | grep -v 'grep' | awk '{ print $1}'` && \
            kill -s HUP $pid