# fish completion for kcp
# Use 'command kcp' to avoid interactions for aliases from kcp to (e.g.) hub

function __fish_kcp_needs_command
	set cmd (commandline -opc)
	if [ (count $cmd) -eq 1 -a $cmd[1] = 'kcp' ]
		return 0
	end
	return 1
end

function __fish_kcp_using_command
	set cmd (commandline -opc)
	if [ (count $cmd) -eq $argv[1] ]
		for i in $argv[3..(count $argv)]
			if [ $i = $cmd[$argv[2]] ]
				return 0
			end
		end
	end
	return 1
end

function __fish_kcp_listall
	set -l appf '/tmp/kcp.lst'
	if not [ -f $appf ]
		command kcp --list-all > $appf
	end
	command cat $appf | sort
end

# General options
complete -fA -c kcp -n '__fish_kcp_needs_command' -a '-h --help'    -d 'Display help'
complete -fA -c kcp -n '__fish_kcp_needs_command' -a '-v --version' -d 'Display version'

# Option get
complete -f -c kcp -n '__fish_kcp_needs_command'              -a '-g --get'             -d 'Download package'
complete -f -c kcp -n '__fish_kcp_using_command 2 2 -g --get' -a '(__fish_kcp_listall)' -d 'Available packages'

# Option install
complete -f -c kcp -n '__fish_kcp_needs_command'                  -a '-i --install'         -d 'Install package'
complete -f -c kcp -n '__fish_kcp_using_command 3 2 -i --install' -a '--asdeps'             -d 'Install package as dependence'
complete -f -c kcp -n '__fish_kcp_using_command 2 2 -i --install' -a '(__fish_kcp_listall)' -d 'Available packages'
complete -f -c kcp -n '__fish_kcp_using_command 3 3 -i --install' -a '(__fish_kcp_listall)' -d 'Available packages'

# Option search
complete -f -c kcp -n '__fish_kcp_needs_command'                 -a '-s --search'          -d 'Search package'
complete -f -c kcp -n '__fish_kcp_using_command 3 2 -s --search' -a '--fast'               -d 'Accelerate search package'
complete -f -c kcp -n '__fish_kcp_using_command 2 2 -s --search' -a '(__fish_kcp_listall)' -d 'Available packages'
complete -f -c kcp -n '__fish_kcp_using_command 3 3 -s --search' -a '(__fish_kcp_listall)' -d 'Available packages'

# Option asdeps
complete -f -c kcp -n '__fish_kcp_needs_command'              -a '--asdeps'  -d 'Install package as dependence'
complete -f -c kcp -n '__fish_kcp_using_command 2 2 --asdeps' -a '--install' -d 'Install package'

# Option fast
complete -f -c kcp -n '__fish_kcp_needs_command'            -a '--fast'   -d 'Accelerate search package'
complete -f -c kcp -n '__fish_kcp_using_command 2 2 --fast' -a '--search' -d 'Search package'
