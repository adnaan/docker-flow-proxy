#!/usr/bin/env python

import subprocess

services = [{'port':'1111','name':'service1'},
          {'port':'2222','name':'service2'},
          {'port':'3333','name':'service3'},
          {'port':'4444','name':'service4'},
          {'port':'5555','name':'service5'}
          ]

customServices = [
          {'port':'2222','name':'service2'},
          {'port':'3333','name':'service3'},
          {'port':'4444','name':'service4'}
          ]

branches = ['master','integration']
subnets  = ['12.0.0.0/24','13.0.0.0/24']

print "Create Main Networks"
for i, b in enumerate(branches):
    subprocess.check_call(['/vagrant/replier/create_overlay.sh', subnets[i], b])

print "Create master and integration containers"

for s in services:
    for b in branches:
        #create containers
        subprocess.check_call(['/vagrant/replier/create_container.sh', s['port'], s['name'],b])

print "Create Custom overlay "
subprocess.check_call(['/vagrant/replier/create_overlay.sh', '14.0.0.0/24', 'custom'])

print "Create custom containers"
for s in customServices:
    #create custom containers
    subprocess.check_call(['/vagrant/replier/create_container.sh', s['port'], s['name'],'custom'])
