#!/usr/bin/env python3
# License: Public Domain
# Release: 0.8

import argparse, os, sys, subprocess, json
from urllib import request

search_head  = 'application/vnd.github.v3.text-match+json'
search_h_tp  = 'Accept'
search_base  = 'https://api.github.com/search/repositories?q={}+user:KaOS-Community-Packages'
url_base     = 'https://github.com/KaOS-Community-Packages/{}.git'
url_pkgbuild = 'https://raw.githubusercontent.com/KaOS-Community-Packages/{}/master/PKGBUILD'

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

def launch_request(search, header = None):
	req = request.Request(search)
	if header:
		req.add_header(search_h_tp, header)
	try:
		result = request.urlopen(req)
		return result.read().decode()
	except:
		return ''

def get_version(result):
	pkgver, pkgrel = '', ''
	for l in result.split('\n'):
		e = l.strip()
		if e[:7] == 'pkgver=':
			pkgver = e[7:]
		elif e[:7] == 'pkgrel=':
			pkgrel = e[7:]
	if pkgver and pkgrel:
		return '{}-{}'.format(pkgver, pkgrel)
	return '<unknown>'

def check_installed(app):
	exe = subprocess.Popen(['pacman', '-Q', app], stdout=subprocess.PIPE, stderr=subprocess.PIPE)	
	if exe.wait():
		return ''
	return exe.stdout.read().decode().split()[1]

def search_package(app, fast):
	search = search_base.format(app)
	result = json.loads(launch_request(search, search_head))
	for a in result['items']:
		n, d, s = a['name'], a['description'], a['stargazers_count']
		i = check_installed(a['name'])
		if fast:
			if i:
				print('\033[1m{}\033[m \033[1;36m[installed: {}]\033[m \033[1;34m({})\033[m'.format(n, i, s))
			else:
				print('\033[1m{}\033[m\033 \033[1;34m({})\033[m'.format(n, s))
		else:
			v = get_version(launch_request(url_pkgbuild.format(n)))
			if i:
				if v == i:
					i = ' [installed]'
				else:
					i = ' [installed: {}]'.format(i)
			print('\033[1m{}\033[m \033[1;32m{}\033[m\033[1;36m{}\033[m \033[1;34m({})\033[m'.format(n, v, i, s))
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
	parser.add_argument('-v', '--version', help='print version', action='version', version='0.8')
	group = parser.add_mutually_exclusive_group()
	group.add_argument('-g', '--get', help='get needed files to build app', metavar='APP')
	group.add_argument('-s', '--search', help='search an app in KCP', metavar='APP')
	group.add_argument('-i', '--install', help='install an app in KCP', metavar='APP')
	parser.add_argument('--asdeps', help='install as a dependence', action='store_true', default=False)
	parser.add_argument('--fast', help='search without version', action='store_true', default=False)
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
	search_package(args.search, args.fast)
elif args.install:
	install_package(args.install, args.asdeps)
else:
	parser.print_help()
