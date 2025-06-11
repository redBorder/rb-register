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

RBDOMAIN=""
CLOUDDOMAIN=""
INSECURE=0
DNS=0
START=1
DNSF=0

source /usr/lib/redborder/lib/rb_functions.sh

function usage(){
	echo "ERROR: $0 [-u <url>] [-i] [-h] [-d]"
  	echo "    -u <url>: url to connect to"
    echo "    -c <cloud_domain>: specify the cloud domain"
  	echo "    -i: do not validate server cert (insecure)"
  	echo "    -d: add dns entries to /etc/hosts in case it is not resolvable and the url is an ip"
    echo "    -f: add the dns entry even if it is a domain (it will try to resolv the ip address)"
  	echo "    -s: do no start services"
  	echo "    -h -> print this help"
	exit 2
}

# Default values
TYPE="proxy"

while getopts "hu:idsft:c:" opt; do
  case $opt in
    i) INSECURE=1;;
    u) RBDOMAIN=$OPTARG;;
    c) CLOUDDOMAIN=$OPTARG;;
    h) usage;;
    d) DNS=1;;
    f) DNSF=1;;
    s) START=0;;
    t) TYPE=$OPTARG;;
  esac
done

if [ ! -f /etc/sysconfig/rb-register.default ]; then
cat >/etc/sysconfig/rb-register.default <<EOF
RBDOMAIN="rblive.redborder.com"
URL="https://\${RBDOMAIN}/api/v1/sensors"
TYPE="$TYPE"
SCRIPT="/usr/lib/redborder/bin/rb_register_finish.sh"
OPTIONS=""
EOF
fi

[ ! -f /etc/sysconfig/rb-register ] && cp /etc/sysconfig/rb-register.default /etc/sysconfig/rb-register

if [ "x$RBDOMAIN" == "x" ]; then
  source /etc/sysconfig/rb-register.default
  [ "x$RBDOMAIN" == "x" ] && RBDOMAIN="rblive.redborder.com"
fi

[ -f /etc/chef/client.rb.default ] && sed -i "s|^chef_server_url.*|chef_server_url  \"https://erchef.service.$CLOUDDOMAIN/organizations/redborder\"|" /etc/chef/client.rb.default
[ -f /etc/chef/client.rb ] && sed -i "s|^chef_server_url.*|chef_server_url  \"https://erchef.service.$CLOUDDOMAIN/organizations/redborder\"|" /etc/chef/client.rb

[ -f /etc/chef/knife.rb.default ] && sed -i "s|^chef_server_url.*|chef_server_url  \"https://erchef.service.$CLOUDDOMAIN/organizations/redborder\"|" /etc/chef/knife.rb.default
[ -f /etc/chef/knife.rb ] && sed -i "s|^chef_server_url.*|chef_server_url  \"https://erchef.service.$CLOUDDOMAIN/organizations/redborder\"|" /etc/chef/knife.rb

[ -f /etc/sysconfig/rb-register.default ] && sed -i "s|^RBDOMAIN=.*|RBDOMAIN=\"$RBDOMAIN\"|" /etc/sysconfig/rb-register.default
[ -f /etc/sysconfig/rb-register ] && sed -i "s|^RBDOMAIN=.*|RBDOMAIN=\"$RBDOMAIN\"|" /etc/sysconfig/rb-register

[ -f /etc/sysconfig/rb-register.default ] && sed -i "s|^URL=.*|URL=\"https://$RBDOMAIN/api/v1/sensors\"|" /etc/sysconfig/rb-register.default
[ -f /etc/sysconfig/rb-register ] && sed -i "s|^URL=.*|URL=\"https://$RBDOMAIN/api/v1/sensors\"|" /etc/sysconfig/rb-register

[ -f /etc/issue ] && sed -i "s|^.*Claim this sensor at.*|NOTE: Claim this sensor at https://$RBDOMAIN with this UUID|" /etc/issue

 
if [ $INSECURE -eq 0 ]; then
  #we must remove ssl_verify_mode entry
  [ -f /etc/chef/client.rb.default ] && sed -i '/^ssl_verify_mode/d' /etc/chef/client.rb.default
  [ -f /etc/chef/client.rb ] && sed -i '/^ssl_verify_mode/d' /etc/chef/client.rb
  [ -f /etc/sysconfig/rb-register.default ] && sed -i 's/-no-check-certificate//' /etc/sysconfig/rb-register.default
  [ -f /etc/sysconfig/rb-register ] && sed -i 's/-no-check-certificate//' /etc/sysconfig/rb-register
else
  for n in /etc/chef/client.rb.default /etc/chef/client.rb; do
    if [ -f $n ]; then
      sed -i '/^ssl_verify_mode/d' $n
      echo "ssl_verify_mode  :verify_none" >> $n
    fi
  done
  for n in /etc/sysconfig/rb-register.default /etc/sysconfig/rb-register; do
    if [ -f $n ]; then
      sed -i 's/-no-check-certificate//' $n
      sed -i 's/^OPTIONS="/OPTIONS="-no-check-certificate /' $n
    fi
  done
fi

#if [ $DNS -eq 1 ]; then
RBDOMAINIP=""
if valid_ip $RBDOMAIN; then
RBDOMAINIP="$RBDOMAIN"
else # [ $DNSF -eq 1 ]; then
RBDOMAINIP=$(getent ahosts $RBDOMAIN 2>/dev/null| awk '{ print $1 }' | head -n 1)
[ "x$RBDOMAINIP" == "x" ] && echo "Cannot get IP from $RBDOMAIN to include it on /etc/hosts, please set a correct DNS"
fi

sed -i '/data.redborder.cluster/d' /etc/hosts
sed -i '/rbookshelf.s3.redborder.cluster/d' /etc/hosts
[ "x$RBDOMAINIP" != "x" ] && echo "$RBDOMAINIP data.$CLOUDDOMAIN $CLOUDDOMAIN s3.service.$CLOUDDOMAIN erchef.service erchef.$CLOUDDOMAIN erchef.service.$CLOUDDOMAIN http2k.service http2k.$CLOUDDOMAIN webui.service" >> /etc/hosts

sed -i '/kafka.service/d' /etc/hosts
echo "127.0.0.1 kafka.service zookeeper.service f2k.service logstash.service freeradius.service n2klocd.service redborder-ale.service rb-nmsp.service rsyslog.service" >> /etc/hosts
#fi

if [ $START -eq 1 ]; then
  if [ -f /etc/chef/client.pem ]; then
    systemctl status chef-client &>/dev/null
    systemctl restart chef-client
  else
    systemctl restart rb-register &>/dev/null
    if [ $? -eq 0 ]; then
      systemctl stop rb-register
    fi
    rm -f /etc/rb-register.db
    systemctl start rb-register
  fi
fi

echo    "Domain to connect: https://$RBDOMAIN"
echo -n "Verify remote certificate: "
if [ $INSECURE -eq 1 ]; then
  echo "disabled"
else
  echo "enabled"
fi