# fish completion for kcp
# Use 'command kcp' to avoid interactions for aliases from kcp to (e.g.) hub

function __fish_kcp_match
	for e in $argv[2..(count $argv)]
		switch $argv[1]
			case "--$e"
				return 0
			case "--*"
				continue
			case "-*$e*"
				return 0
		end
	end
	return 1
end

function __fish_kcp_match_last
	for e in $argv[2..(count $argv)]
		switch $argv[1]
			case "--$e"
				return 0
			case "--*"
				continue
			case "-*$e"
				return 0
		end
	end
	return 1
end

function __fish_kcp_contains
	for e in $argv[3..(count $argv)]
		if __fish_kcp_match $e $argv[1..2]
			return 0
		end
	end
	return 1
end

function __fish_kcp_contains_last
	for e in $argv[3..(count $argv)]
		if __fish_kcp_match_last $e $argv[1..2]
			return 0
		end
	end
	return 1
end

function __fish_kcp_needs_arg
	set cmd (commandline -opc)
	if __fish_kcp_match_last $cmd[(count $cmd)] s i g V search install get information
		return 0
	end
	return 1
end

function __fish_kcp_empty
	set cmd (commandline -opc)
	if [ (count $cmd) -eq 1 -a $cmd[1] = 'kcp' ]
		return 0
	end
	return 1
end

function __fish_kcp_needs_command
	if __fish_kcp_needs_arg
		return 1
	end
	set cmd (commandline -opc)
	if __fish_kcp_contains $argv $cmd
		return 1
	end
	switch $argv[1]
		case l list
			if __fish_kcp_contains s search $cmd
				return 1
			end
			if __fish_kcp_contains u update-database $cmd
				return 1
			end
			if __fish_kcp_contains N only-name $cmd
				return 0
			end
			if __fish_kcp_contains S only-starred $cmd
				return 0
			end
			if __fish_kcp_contains I only-installed $cmd
				return 0
			end
			if __fish_kcp_contains O only-outdated $cmd
				return 0
			end
			if __fish_kcp_contains x sort $cmd
				return 0
			end
			if __fish_kcp_contains f force-update $cmd
				return 0
			end
		case u update-database
			if __fish_kcp_contains l list $cmd
				return 1
			end
		case s search
			if __fish_kcp_contains l list $cmd
				return 1
			end
			if __fish_kcp_contains N only-name $cmd
				return 0
			end
			if __fish_kcp_contains S only-starred $cmd
				return 0
			end
			if __fish_kcp_contains I only-installed $cmd
				return 0
			end
			if __fish_kcp_contains O only-outdated $cmd
				return 0
			end
			if __fish_kcp_contains x sort $cmd
				return 0
			end
			if __fish_kcp_contains f force-update $cmd
				return 0
			end
		case i install
			if __fish_kcp_contains d asdeps $cmd
				return 0
			end
		case N only-name S only-starred I only-installed O only-outdated x sort
			if __fish_kcp_contains l list $cmd
				return 0
			end
			if __fish_kcp_contains s search $cmd
				return 0
			end
			if __fish_kcp_contains N only-name $cmd
				return 0
			end
			if __fish_kcp_contains S only-starred $cmd
				return 0
			end
			if __fish_kcp_contains I only-installed $cmd
				return 0
			end
			if __fish_kcp_contains O only-outdated $cmd
				return 0
			end
			if __fish_kcp_contains x sort $cmd
				return 0
			end
			if __fish_kcp_contains f force-update $cmd
				return 0
			end
		case f force-update
			if __fish_kcp_contains u update-database $cmd
				return 1
			end
			if __fish_kcp_contains s search $cmd
				return 0
			end
			if __fish_kcp_contains l list $cmd
				return 0
			end
			if __fish_kcp_contains N only-name $cmd
				return 0
			end
			if __fish_kcp_contains S only-starred $cmd
				return 0
			end
			if __fish_kcp_contains I only-installed $cmd
				return 0
			end
			if __fish_kcp_contains O only-outdated $cmd
				return 0
			end
			if __fish_kcp_contains x sort $cmd
				return 0
			end
		case d asdeps
			if __fish_kcp_contains i install $cmd
				return 0
			end
	end
	return 1
end

function __fish_kcp_listall
	command kcp -lN | sort
end

# Initialization
complete -fA -c kcp -n '__fish_kcp_empty' -a '-h --help'            -d 'Display help'
complete -fA -c kcp -n '__fish_kcp_empty' -a '-v --version'         -d 'Display version'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-l --list'            -d 'List all available packages'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-u --update-database' -d 'Refresh the local database'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-s --search'          -d 'Search a package'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-g --get'             -d 'Download a package'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-i --install'         -d 'Install a package'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-V --information'     -d 'Information about a package'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-lN --only-name'      -d 'Display only the packages name'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-lS --only-starred'   -d 'Display only the popular packages'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-lI --only-installed' -d 'Display only the installed packages'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-lO --only-outdated'  -d 'Display only the outdated packages'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-lx --sort'           -d 'Sort results by popularity'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-lf --force-update'   -d 'Force refreshing database'
complete -f  -c kcp -n '__fish_kcp_empty' -a '-di --asdeps'         -d 'install as dependence'


# Options
complete -f -c kcp -n '__fish_kcp_needs_command l list'            -a '-l --list'            -d 'Display all available packages'
complete -f -c kcp -n '__fish_kcp_needs_command u update-database' -a '-u --update-database' -d 'Refresh the local database'
complete -f -c kcp -n '__fish_kcp_needs_command s search'          -a '-s --search'          -d 'Search a package'
complete -f -c kcp -n '__fish_kcp_needs_command i install'         -a '-i --install'         -d 'Install a package'
complete -f -c kcp -n '__fish_kcp_needs_command N only-name'       -a '-N --only-name'       -d 'Display only the packages name'
complete -f -c kcp -n '__fish_kcp_needs_command S only-starred'    -a '-S --only-starred'    -d 'Display only the popular packages'
complete -f -c kcp -n '__fish_kcp_needs_command I only-installed'  -a '-I --only-installed'  -d 'Display only the installed packages'
complete -f -c kcp -n '__fish_kcp_needs_command O only-outdated'   -a '-O --only-outdated'   -d 'Display only the outdated packages'
complete -f -c kcp -n '__fish_kcp_needs_command x sort'            -a '-x --sort'            -d 'Sort results by popularity'
complete -f -c kcp -n '__fish_kcp_needs_command f force-update'    -a '-f --force-update'    -d 'Force refreshing database'
complete -f -c kcp -n '__fish_kcp_needs_command d asdeps'          -a '-d --asdeps'          -d 'install as dependence'

# Available packages
complete -f -c kcp -n '__fish_kcp_needs_arg' -a '(__fish_kcp_listall)' -d 'Available packages'
