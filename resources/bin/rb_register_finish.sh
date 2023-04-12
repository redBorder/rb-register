#!/bin/bash

#######################################################################
# Copyright (c) 2014 ENEO Tecnología S.L.
# This file is part of redBorder.
# redBorder is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# redBorder is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License License for more details.
# You should have received a copy of the GNU Affero General Public License License
# along with redBorder. If not, see <http://www.gnu.org/licenses/>.
#######################################################################

source /etc/profile

sensor_id="0"
counter=1
max=10
ret=1

function title(){
  echo "############################################################################"
  echo "$*"
  echo "############################################################################"
}

if [ -f /etc/chef/role-once.json.default ]; then
  title "  started rb_register_finish.sh ($(date))"

  #touch /etc/force_create_topics

  rm -f /etc/chef/role.json /etc/chef/role-once.json /etc/rb-id /etc/chef/client.rb /root/.chef/knife.rb

  cp /etc/chef/role-once.json.default /etc/chef/role-once.json
  
  if [ -f /etc/chef/client.rb.default ]; then
    cp -f /etc/chef/client.rb.default /etc/chef/client.rb
  else
    echo "This node has not chef well configure!! (/etc/chef/client.rb.default)"
    exit 1 
  fi

  if [ -f /etc/chef/knife.rb.default ]; then
    mkdir -p /root/.chef
    cp -f /etc/chef/knife.rb.default /root/.chef/knife.rb
  else
    echo "This node has not chef well configure!! (/etc/chef/knife.rb.default)"
    exit 1 
  fi

  if [ -f /etc/chef/nodename ]; then
    NODENAME=$(cat /etc/chef/nodename)
    [ -f /etc/chef/client.rb ] && sed -i "s|HOSTNAME|$NODENAME|" /etc/chef/client.rb
    [ -f /root/.chef/knife.rb ] && sed -i "s|HOSTNAME|$NODENAME|" /root/.chef/knife.rb
    [ -f /etc/hosts ] && sed -i "s|kafka.service|$NODENAME.node kafka.service|" /etc/hosts
  else
    echo "This node has not valid nodename yet!! (/etc/chef/nodename)"
    exit 1 
  fi

  # Set hostname
  hostnamectl set-hostname $NODENAME

  while [ "x$sensor_id" == "x0" -a $counter -le $max -a ! -f /etc/chef/role.json ]; do
    title "       chef-client run (${counter})"
    
    #########################
    # rb_run_chef_once.sh   #
    #########################
    #chef-client -c /etc/chef/client.rb --once -s 5 --node-name $(head -n 1 /etc/chef/nodename) -j /etc/chef/role-once.json
    chef-client -c /etc/chef/client.rb --once -s 5 -j /etc/chef/role-once.json

    sensor_id=$(head -n 1 /etc/rb-id 2>/dev/null)
    counter=$(($counter +1))
    [ "x$sensor_id" == "x" ] && sensor_id=0
    [ $sensor_id -eq 0 -a $counter -lt 10 ] && sleep 10
  done

  [ -f /etc/chef/role.json -a "x$sensor_id" != "x" ] && ret=0
  [ ! -f /etc/chef/role.json ] && cp /etc/chef/role.json.default /etc/chef/role.json

  #service zookeeper status &>/dev/null
  #[ $? -eq 0 ] && timeout 300 /opt/rb/bin/rb_create_topics.sh | grep -v 'Due to limitations in metric names' | grep -v "already exists" | grep -v "kafka.admin"

  #rm -f /etc/force_create_topics

  if [ -f /etc/chef/client.pem ]; then
    systemctl enable chef-client
    # systemctl disable rb-register
    systemctl stop rb-register
    systemctl start chef-client
    sleep 5
    rb_wakeup_chef.sh
  fi
 
  title "  finished rb_register_finish.sh ($(date))"
  date > /etc/redborder/sensor-installed.txt
else
  echo "ERROR: /etc/chef/role-once.json.default not found"
fi

exit $ret
