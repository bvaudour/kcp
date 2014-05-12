#!/usr/bin/env python3
# License: Public Domain
# Release: 0.5

import argparse, os, sys, subprocess, json
from urllib import request

search_head = 'application/vnd.github.v3.text-match+json'
search_h_tp = 'Accept'
search_base = 'https://api.github.com/search/repositories?q={}+user:KaOS-Community-Packages'
url_base    = 'https://github.com/KaOS-Community-Packages/{}.git'

def print_error(msg):
	print('\033[1;31m{}\033[m'.format(msg))

def question(msg, default_value=True):
	if default_value:
		choice_rep = '[Y/n]'
	else:
		choice_rep = '[y/N]'
	response = input('\033[1;33m{} {} \033[m'.format(msg, choice_rep)).lower()
	if not response:
		return default_value
	if response[0] == 'y':
		return True
	if response[0] == 'n':
		return False
	return default_value

def edit(file_name):
	os.system('$EDITOR {}'.format(file_name))

def check_user():
	if not os.geteuid():
		print_error("Don't launch this program as root!")
		sys.exit(1)

def edit_pkgbuild():
	if question('Do you want to edit PKGBUILD?'):
		edit('PKGBUILD')

def get_package(app):
	url = url_base.format(app)
	exe = subprocess.Popen(['git', 'clone', url])
	err = exe.wait()
	if err:
		sys.exit(err)

def launch_request(search):
	req = request.Request(search)
	req.add_header(search_h_tp, search_head)
	return request.urlopen(req).read().decode()

def search_package(app):
	search = search_base.format(app)
	result = json.loads(launch_request(search))
	for a in result['items']:
		n, d, s = a['name'], a['description'], a['stargazers_count']
		print('\033[1m{}\033[m \033[1;36m({})\033[m'.format(n, s))
		print('\t{}'.format(d))

def install_package(app, asdeps):
	os.chdir('/tmp')
	get_package(app)
	os.chdir('/tmp/{}'.format(app))
	edit_pkgbuild()
	cmd = 'makepkg -si'
	if asdeps:
		cmd += ' --asdeps'
	os.system(cmd)
	os.system('rm -rf {}'.format(os.getcwd()))

def build_args():
	parser = argparse.ArgumentParser(description='Tool in command-line for KaOS Community Packages')
	parser.add_argument('-v', '--version', help='print version', action='version', version='0.5')
	group = parser.add_mutually_exclusive_group()
	group.add_argument('-g', '--get', help='get needed files to build app', metavar='APP')
	group.add_argument('-s', '--search', help='search an app in KCP', metavar='APP')
	group.add_argument('-i', '--install', help='install an app in KCP', metavar='APP')
	parser.add_argument('--asdeps', help='install as a dependence', action='store_true', default=False)
	return parser

check_user()

parser = build_args()
if len(sys.argv) == 1:
	parser.print_help()
	sys.exit(0)

args = parser.parse_args()
if args.get:
	get_package(args.get)
elif args.search:
	search_package(args.search)
elif args.install:
	install_package(args.install, args.asdeps)
else:
	parser.print_help()
