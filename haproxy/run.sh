#!/usr/bin/env bash

set -e

if [ -f /var/run/haproxy.pid ]; then
    SERVICE_NAME=$1
    SERVICE_PATH=$2

    echo "
    frontend ${SERVICE_NAME}-fe
        bind *:80
        bind *:443
        option http-server-close
        acl url_${SERVICE_NAME} path_beg ${SERVICE_PATH}
        use_backend ${SERVICE_NAME}-be if url_${SERVICE_NAME}

    backend ${SERVICE_NAME}-be
        {{range service \"${SERVICE_NAME}\" \"any\"}}
        server {{.Node}}_{{.Port}} {{.Address}}:{{.Port}} check
        {{end}}
    " >/usr/local/etc/haproxy/tmpl/service-formatted.ctmpl

    consul-template \
        -consul $CONSUL_IP:$CONSUL_PORT \
        -template "/usr/local/etc/haproxy/tmpl/service-formatted.ctmpl:/usr/local/etc/haproxy/tmpl/$SERVICE_NAME.cfg" \
        -once

    cat /usr/local/etc/haproxy/tmpl/haproxy.tmpl /usr/local/etc/haproxy/tmpl/*.cfg >/usr/local/etc/haproxy/haproxy.cfg

    haproxy -f /usr/local/etc/haproxy/haproxy.cfg -D -p /var/run/haproxy.pid -sf $(cat /var/run/haproxy.pid)
else
    haproxy -f /usr/local/etc/haproxy/haproxy.cfg -D -p /var/run/haproxy.pid
    sleep infinity
fi
