#!/usr/bin/env python3
# License: Public Domain
# Release: 0.1

import argparse, os, sys, subprocess
from io import BytesIO
from contextlib import contextmanager

search_head = 'Accept: application/vnd.github.v3.text-match+json'
search_base = 'https://api.github.com/search/repositories?q={}+user:KaOS-Community-Packages'
url_base    = 'https://github.com/KaOS-Community-Packages/{}.git'

def print_error(msg):
	print('\033[1;31m{}\033[m'.format(msg))

def capture_output(stdout_to = None, stderr_to = None):
	try:
		stdout, stderr = sys.stdout, sys.stderr
		sys.stdout = c1 = stdout_to or BytesIO()
		sys.stderr = c2 = stderr_to or BytesIO()
		yield c1, c2
	finally:
		sys.stdout = stdout
		sys.stderr = stderr
		try:
			c1.flush(1)
			c1.seek(0)
		except (ValueError, IOError):
			pass
		try:
			c2.flush()
			c2.seek(0)
		except (ValueError, IOError):
			pass

def get_package(app):
	url = url_base.format(app)
	exe = subprocess.Popen(['git', 'clone', url])
	err = exe.wait()
	if err:
		sys.exit(err)

def search_package(app):
	search = search_base.format(app)
	exe = subprocess.Popen(['curl', '-H', search_head, search], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
	err = exe.wait()
	if err:
		print(exe.stderr.read())
		sys.exit(err)
	json = str(exe.stdout.read()).split('\\n')
	(name, description, star) = ([], [], [])
	for l in json:
		e = l.strip()
		n = e.find(':')
		if n < 0:
			continue
		key, value = e[:n], e[n+1:-1].strip()
		if key == '"name"':
			name.append(value[1:-1])
		elif key == '"description"':
			description.append(value[1:-1])
		elif key == '"stargazers_count"':
			star.append(value)
	for i in range(len(name)):
		n, d, s = name[i], description[i], star[i]
		print('\033[1m{}\033[m \033[1;36m({})\033[m'.format(n, s))
		print('\t{}'.format(d))

def install_package(app, asdeps):
	os.chdir('/tmp')
	get_package(app)
	os.chdir('/tmp/{}'.format(app))
	cmd = 'makepkg -si'
	if asdeps:
		cmd += ' --asdeps'
	os.system(cmd)
	os.system('rm -rf {}'.format(os.getcwd()))

def build_args():
	parser = argparse.ArgumentParser(description='Tool in command-line for KaOS Community Packages')
	parser.add_argument('-v', '--version', help='print version', action='version', version='0.1')
	group = parser.add_mutually_exclusive_group()
	group.add_argument('-g', '--get', help='get needed files to build app', metavar='APP')
	group.add_argument('-s', '--search', help='search an app in KCP', metavar='APP')
	group.add_argument('-i', '--install', help='install an app in KCP', metavar='APP')
	parser.add_argument('--asdeps', help='install as a dependence', action='store_true', default=False)
	return parser

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
