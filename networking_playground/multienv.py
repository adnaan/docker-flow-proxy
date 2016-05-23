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
branches_nonet = ['master-nonet','integration-nonet']
subnets  = ['12.0.0.0/24','13.0.0.0/24']
subnets_nonet = ['14.0.0.0/24','15.0.0.0/24']

print "Create Main Networks"
for i, b in enumerate(branches):
    subprocess.check_call(['/vagrant/networking_playground/create_overlay.sh', subnets[i], b])

print "Create master and integration containers"

for s in services:
    for b in branches:
        #create containers
        subprocess.check_call(['/vagrant/networking_playground/create_container.sh', s['port'], s['name'],b,'true'])

print "Create Main Networks for nonet containers"
for i, b in enumerate(branches_nonet):
    subprocess.check_call(['/vagrant/networking_playground/create_overlay.sh', subnets_nonet[i], b])

print "Create master-nonet and integration-nonet containers"

for s in services:
    for b in branches_nonet:
        #create containers
        subprocess.check_call(['/vagrant/networking_playground/create_container.sh', s['port'], s['name'],b,'false'])

print "Create Custom overlay "
subprocess.check_call(['/vagrant/networking_playground/create_overlay.sh', '14.0.0.0/24', 'custom'])

print "Create custom containers"
for s in customServices:
    #create custom containers
    subprocess.check_call(['/vagrant/networking_playground/create_container.sh', s['port'], s['name'],'custom','true'])

print "Created!"
