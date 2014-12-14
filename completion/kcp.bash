# kcp completion

_kcpContains() {
	for e in "${@:2}"; do
		[[ "$e" == "$1" ]] && return 0
	done
  return 1
}

_kcpContains2() {
	_kcpContains $1 ${@:3} && return 0
	_kcpContains $2 ${@:3} && return 0
	return 1
}

_kcpNoSearch() {
	_kcpContains2 "-s" "--search" $@ && return 1
	return 0
}

_kcpNoInstall() {
	_kcpContains2 "-i" "--install" $@ && return 1
	return 0
}

_kcpNoList() {
	_kcpContains2 "-l" "--list" $@ && return 1
	return 0
}

_kcpNoOut() {
	_kcpContains2 "-o" "--outdated" $@ && return 1
	return 0
}

_kcpNoAsdeps() {
	_kcpContains "--asdeps" $@ && return 1
	return 0
}

_kcpNoFast() {
	_kcpContains "--fast" $@ && return 1
	return 0
}

_kcpNoSort() {
	_kcpContains "--sort" $@ && return 1
	return 0
}

_kcp() {
	local cur prev words cword pprev opts lst appfile appcache
	_init_completion || return
	pprev="${COMP_WORDS[COMP_CWORD-2]}"
	opts="--asdeps --fast --sort --help --install --search --get --version --outdated --list -h -i -s -g -v -o -l"
	appfile="/tmp/kcp.lst"

	case "$prev" in
		"kcp")
			lst="$opts"
			;;
		"--help"|"--version"|"-h"|"-v")
			return 0
			;;
		"--install"|"--search"|"--get"|"-i"|"-s"|"-g")
			[[ -f "$appfile" ]] || kcp -l --fast > "$appfile"
			lst=$( cat "$appfile" | sort )
			;;
		"-o"|"--outdated")
			_kcpNoSort ${COMP_WORDS[@]} && lst="--sort"
			;;
		"-l"|"--list")
			_kcpNoSort ${COMP_WORDS[@]} && lst="--sort"
			_kcpNoFast ${COMP_WORDS[@]} && lst="$lst --fast"
			;;
		"--asdeps")
			_kcpNoInstall ${COMP_WORDS[@]} && lst="--install"
			;;
		"--fast")
			_kcpNoSearch ${COMP_WORDS[@]} && _kcpNoList ${COMP_WORDS[@]} && lst="--search --list"
			_kcpNoSort ${COMP_WORDS[@]} && lst="$lst --sort"
			;;
		"--sort")
			if _kcpNoOut ${COMP_WORDS[@]}; then
				_kcpNoSearch ${COMP_WORDS[@]} && _kcpNoList ${COMP_WORDS[@]} && lst="--search --list --outdated"
				_kcpNoFast ${COMP_WORDS[@]} && lst="$lst --fast"
			fi
			;;
		*)
			case "$pprev" in
				"--install"|"-i")
					_kcpNoAsdeps ${COMP_WORDS[@]} && lst="--asdeps"
					;;
				"--search"|"-s")
					_kcpNoSort ${COMP_WORDS[@]} && lst="--sort"
					_kcpNoFast ${COMP_WORDS[@]} && lst="$lst --fast"
					;;
				*)
					return 0
					;;
			esac
			;;
	esac
	
	[[ $lst == "" ]] && return 0
	COMPREPLY=( $(compgen -W "$lst" -- "$cur") )
}

complete -F _kcp kcp
