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

print "Remove All Containers"
for s in services:
    for b in branches:
        try:
            subprocess.check_call(['/vagrant/replier/remove_container.sh', s['name'],b])
        except:
            pass



for s in customServices:
    for b in branches:
        try:
            subprocess.check_call(['/vagrant/replier/remove_container.sh', s['name'],'custom'])
        except:
            pass

print "Remove All Images"
for s in services:
    for b in branches:
        try:
            subprocess.check_call(['/vagrant/replier/remove_image.sh', s['name'],b])
        except:
            pass

for s in customServices:
    for b in branches:
        try:
            subprocess.check_call(['/vagrant/replier/remove_image.sh', s['name'],'custom'])
        except:
            pass

print "Remove Main Networks"
for i, b in enumerate(branches):
    try:
        subprocess.check_call(['/vagrant/replier/remove_overlay.sh', b])
    except:
        pass

print "Remove Custom Network"
try:
    subprocess.check_call(['/vagrant/replier/remove_overlay.sh', 'custom'])
except:
    pass
