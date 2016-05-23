#!/usr/bin/env python

import subprocess

services = [{'port':'1111','name':'service1'},
          {'port':'2222','name':'service2'},
          {'port':'3333','name':'service3'},
          {'port':'4444','name':'service4'},
          {'port':'5555','name':'service5'}
          ]

customServices = [
          {'port':'4444','name':'service4'},
          {'port':'3333','name':'service3'},
          {'port':'2222','name':'service2'}
          ]

branches = ['master','integration']
branches_nonet = ['master-nonet','integration-nonet']

print "Remove All Containers"
for s in services:
    for b in branches:
        subprocess.check_call(['/vagrant/replier/remove_container.sh', s['name'],b])


for s in services:
    for b in branches_nonet:
        subprocess.check_call(['/vagrant/replier/remove_container.sh', s['name'],b])



for s in customServices:
    subprocess.check_call(['/vagrant/replier/remove_container.sh', s['name'],'custom'])

print "Remove All Images"
for s in services:
    for b in branches:
        subprocess.check_call(['/vagrant/replier/remove_image.sh', s['name'],b])

for s in services:
    for b in branches_nonet:
        subprocess.check_call(['/vagrant/replier/remove_image.sh', s['name'],b])

for s in customServices:
    subprocess.check_call(['/vagrant/replier/remove_image.sh', s['name'],'custom'])

print "Remove Main Networks"
for i, b in enumerate(branches):
    subprocess.check_call(['/vagrant/replier/remove_overlay.sh', b])

for i, b in enumerate(branches_nonet):
    subprocess.check_call(['/vagrant/replier/remove_overlay.sh', b])

print "Remove Custom Network"
subprocess.check_call(['/vagrant/replier/remove_overlay.sh', 'custom'])

print "Destroyed!"
