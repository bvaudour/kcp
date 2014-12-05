# kcp completion

_kcp() {
	local cur prev words cword pprev opts lst appfile appcache
	_init_completion || return
	pprev="${COMP_WORDS[COMP_CWORD-2]}"
	opts="--asdeps --fast --help --install --search --get --version -h -i -s -g -v"
	appfile="/tmp/kcp.lst"

	case "$prev" in
		"kcp")
			lst="$opts"
			;;
		"--asdeps")
			lst="--install"
			;;
		"--fast")
			lst="--search"
			;;
		"--install"|"--search"|"--get"|"-i"|"-s"|"-g")
			if [[ ! -f "$appfile" ]]; then
				kcp --list-all > "$appfile"
			fi
			lst=$( cat "$appfile" | sort )
			;;
		"--help"|"--version"|"-h"|"-v")
			return 0
			;;
		*)
			case "$pprev" in
				"--install"|"-i")
					lst="--asdeps"
					;;
				"--search"|"-s")
					lst="--fast"
					;;
				*)
					return 0
					;;
			esac
	esac

	COMPREPLY=( $(compgen -W "$lst" -- "$cur") )
}

complete -F _kcp kcp
