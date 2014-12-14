# fish completion for kcp
# Use 'command kcp' to avoid interactions for aliases from kcp to (e.g.) hub

function __fish_kcp_needs_arg
	set cmd (commandline -opc)
	switch $cmd[(count $cmd)]
		case -s -i -g --search --install --get
			return 0
		case '*'
			return 1
	end
end

function __fish_kcp_empty
	set cmd (commandline -opc)
	if [ (count $cmd) -eq 1 -a $cmd[1] = 'kcp' ]
		return 0
	end
	return 1
end

function __fish_kcp_command_used
	set cmd (commandline -opc)
	for c in $cmd
		switch $c
			case $argv
				return 0
		end
	end
	return 1
end

function __fish_kcp_needs_command
	if __fish_kcp_empty
		return 0
	end
	if __fish_kcp_needs_arg
		return 1
	end
	if __fish_kcp_command_used $argv
		return 1
	end
	switch $argv[1]
		case -i --install
			if __fish_kcp_command_used --asdeps
				return 0
			end
		case --asdeps
			if __fish_kcp_command_used -i --install
				return 0
			end
			return 1
		case -s --search
			if __fish_kcp_command_used -l -o --list --outdated
				return 1
			end
			if __fish_kcp_command_used --sort --fast
				return 0
			end
		case -l --list
			if __fish_kcp_command_used -s -o --search --outdated
				return 1
			end
			if __fish_kcp_command_used --sort --fast
				return 0
			end
		case -o --outdated
			if __fish_kcp_command_used -s -l --search --list
				return 1
			end
			if __fish_kcp_command_used --sort
				return 0
			end
		case --sort
			if __fish_kcp_command_used -s -l -o --search --list --outdated --fast
				return 0
			end
		case --fast
			if __fish_kcp_command_used -o --outdated
				return 1
			end
			if __fish_kcp_command_used -s -l --search --list --sort
				return 0
			end
	end
	return 1
end

function __fish_kcp_listall
	set -l appf '/tmp/kcp.lst'
	if not [ -f $appf ]
		command kcp -l --fast > $appf
	end
	command cat $appf | sort
end

# General Options
complete -fA -c kcp -n '__fish_kcp_needs_command -h --help'    -a '-h --help'    -d 'Display help'
complete -fA -c kcp -n '__fish_kcp_needs_command -v --version' -a '-v --version' -d 'Display version'

# Required options
complete -f -c kcp -n '__fish_kcp_needs_command -l --list'     -a '-l --list'     -d 'Display all packages'
complete -f -c kcp -n '__fish_kcp_needs_command -o --outdated' -a '-o --outdated' -d 'Display outdated packages'
complete -f -c kcp -n '__fish_kcp_needs_command -s --search'   -a '-s --search'   -d 'Search package'
complete -f -c kcp -n '__fish_kcp_needs_command -i --install'  -a '-i --install'  -d 'Install package'
complete -f -c kcp -n '__fish_kcp_needs_command -g --get'      -a '-g --get'      -d 'Download package'

# Not required options
complete -f -c kcp -n '__fish_kcp_needs_command --fast'   -a '--fast'   -d 'Not display version'
complete -f -c kcp -n '__fish_kcp_needs_command --sort'   -a '--sort'   -d 'Sort results by stars descending'
complete -f -c kcp -n '__fish_kcp_needs_command --asdeps' -a '--asdeps' -d 'Install package as dependence'

# Available packages
complete -f -c kcp -n '__fish_kcp_needs_arg' -a '(__fish_kcp_listall)' -d 'Available packages'
