# kcp completion

_kcp() {
	local cur prev words cword pprev opts lst appfile appcache
	_init_completion || return
	pprev="${COMP_WORDS[COMP_CWORD-2]}"
	opts="--asdeps --fast --stars --help --install --search --get --version --outdated -h -i -s -g -v -o"
	appfile="/tmp/kcp.lst"

	case "$prev" in
		"kcp")
			lst="$opts"
			;;
		"--asdeps")
			lst="--install"
			;;
		"--fast"|"--stars")
			lst="--search"
			;;
		"--install"|"--search"|"--get"|"-i"|"-s"|"-g")
			if [[ ! -f "$appfile" ]]; then
				kcp --list-all > "$appfile"
			fi
			lst=$( cat "$appfile" | sort )
			;;
		"--help"|"--version"|"--outdated"|"-h"|"-v"|'-o')
			return 0
			;;
		*)
			case "$pprev" in
				"--install"|"-i")
					lst="--asdeps"
					;;
				"--search"|"-s")
					lst="--fast --stars"
					;;
				*)
					return 0
					;;
			esac
	esac

	COMPREPLY=( $(compgen -W "$lst" -- "$cur") )
}

complete -F _kcp kcp
