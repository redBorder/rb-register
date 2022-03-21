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
INSECURE=0
DNS=0
START=1
DNSF=0

source /usr/lib/redborder/lib/rb_proxy_functions.sh

function usage(){
	echo "ERROR: $0 [-u <url>] [-i] [-h] [-d]"
  	echo "    -u <url>: url to connect to"
  	echo "    -i: do not validate server cert (insecure)"
  	echo "    -d: add dns entries to /etc/hosts in case it is not resolvable and the url is an ip"
        echo "    -f: add the dns entry even if it is a domain (it will try to resolv the ip address)"
  	echo "    -s: do no start services"
  	echo "    -h -> print this help"
	exit 2
}

while getopts "hu:idsf" opt; do
  case $opt in
    i) INSECURE=1;;
    u) RBDOMAIN=$OPTARG;;
    h) usage;;
    d) DNS=1;;
    f) DNSF=1;;
    s) START=0;;
  esac
done

if [ "x$RBDOMAIN" == "x" ]; then
  source /etc/sysconfig/rb-register.default
  [ "x$RBDOMAIN" == "x" ] && RBDOMAIN="rblive.redborder.com"
fi

[ -f /etc/chef/client.rb.default ] && sed -i "s|^chef_server_url.*|chef_server_url  \"https://$RBDOMAIN\"|" /etc/chef/client.rb.default
[ -f /etc/chef/client.rb ] && sed -i "s|^chef_server_url.*|chef_server_url  \"https://$RBDOMAIN\"|" /etc/chef/client.rb
[ -f /etc/sysconfig/rb-register.default ] && sed -i "s|^RBDOMAIN=.*|RBDOMAIN=\"$RBDOMAIN\"|" /etc/sysconfig/rb-register.default
[ -f /etc/sysconfig/rb-register ] && sed -i "s|^RBDOMAIN=.*|RBDOMAIN=\"$RBDOMAIN\"|" /etc/sysconfig/rb-register
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
[ "x$RBDOMAINIP" != "x" ] && echo "$RBDOMAINIP data.redborder.cluster rbookshelf.s3.redborder.cluster redborder.cluster" >> /etc/hosts
#fi

if [ $START -eq 1 ]; then
  if [ -f /etc/chef/client.pem ]; then
    /etc/init.d/chef-client status &>/dev/null
    if [ $? -eq 0 ]; then
      service chef-client restart
    else
      service chef-client restart
    fi
  else
    /etc/init.d/rb-register status &>/dev/null
    if [ $? -eq 0 ]; then
      service rb-register stop
      rm -f /etc/rb-register.db
      service rb-register start
    else
      rm -f /etc/rb-register.db
      service rb-register start
    fi
  fi
fi

echo    "Domain to connect: https://$RBDOMAIN"
echo -n "Verify remote certificate: "
if [ $INSECURE -eq 1 ]; then
  echo "disabled"
else
  echo "enabled"
fi