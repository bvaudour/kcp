# kcp completion

_kcpMatch() {
	for i in ${@:2}; do
		case $1 in
			--$i)  return 0;;
			--*)   ;;
			-*$i*) return 0;;
			*)     ;;
		esac
	done
	return 1
}

_kcpMatchLast() {
	for i in ${@:2}; do
		case $1 in
			--$i)  return 0;;
			--*)   ;;
			-*$i) return 0;;
			*)     ;;
		esac
	done
	return 1
}

_kcpContains() {
	for e in ${@:3}; do
		_kcpMatch $e $1 $2 && return 0
	done
	return 1
}

_kcpContainsLast() {
	for e in ${@:3}; do
		_kcpMatchLast $e $1 $2 && return 0
	done
	return 1
}

_kcpIsInstall() {
	_kcpContainsLast i install $@ && return 0
	return 1
}

_kcpIsUpdate() {
	_kcpContains u update-database $@ && return 0
	return 1
}

_kcpIsSearch() {
	_kcpContainsLast s search $@ && return 0
	return 1
}

_kcpIsList() {
	_kcpContains l list $@ && return 0
	return 1
}

_kcpIsInfo() {
	_kcpContains V information $@ && return 0
	return 1
}

_kcp() {
	local cur prev words cword pprev opts lst
	_init_completion || return
	pprev="${COMP_WORDS[COMP_CWORD-2]}"
	opts='--help --version --list --update-database --search --get --install
				--only-name --only-starred --only-installed --only-outdated
				--sort --force-update --asdeps --information
				-h -v -l -u -s -g -i -V
				-lN -lS -lI -lO
				-lx -lf -di'
	lst=()

	_kcpMatch $prev h help v version && return 0
	if [[ $prev == "kcp" ]]; then
		lst=($opts)
	elif _kcpMatchLast $prev s search g get i install V information; then
		lst=( $( kcp -lN | sort ) )
	elif _kcpIsInstall ${COMP_WORDS[@]}; then
		_kcpContains d asdeps ${COMP_WORDS[@]} || lst=(--asdeps)
	elif _kcpIsSearch ${COMP_WORDS[@]}; then
		_kcpContains x sort ${COMP_WORDS[@]} || lst=(${lst[@]} --sort)
		_kcpContains N only-name ${COMP_WORDS[@]} || lst=(${lst[@]} --only-name)
		_kcpContains S only-starred ${COMP_WORDS[@]} || lst=(${lst[@]} --only-starred)
		_kcpContains I only-installed ${COMP_WORDS[@]} || lst=(${lst[@]} --only-installed)
		_kcpContains O only-outdated ${COMP_WORDS[@]} || lst=(${lst[@]} --only-outdated)
		_kcpContains f force-update ${COMP_WORDS[@]} || lst=(${lst[@]} --force-update)
	elif _kcpIsList ${COMP_WORDS[@]}; then
		_kcpContains x sort ${COMP_WORDS[@]} || lst=(${lst[@]} --sort)
		_kcpContains N only-name ${COMP_WORDS[@]} || lst=(${lst[@]} --only-name)
		_kcpContains S only-starred ${COMP_WORDS[@]} || lst=(${lst[@]} --only-starred)
		_kcpContains I only-installed ${COMP_WORDS[@]} || lst=(${lst[@]} --only-installed)
		_kcpContains O only-outdated ${COMP_WORDS[@]} || lst=(${lst[@]} --only-outdated)
		_kcpContains f force-update ${COMP_WORDS[@]} || lst=(${lst[@]} --force-update)
	elif _kcpContains f force-update ${COMP_WORDS[@]}; then
		lst=(--list --search)
		_kcpContains x sort ${COMP_WORDS[@]} || lst=(${lst[@]} --sort)
		_kcpContains N only-name ${COMP_WORDS[@]} || lst=(${lst[@]} --only-name)
		_kcpContains S only-starred ${COMP_WORDS[@]} || lst=(${lst[@]} --only-starred)
		_kcpContains I only-installed ${COMP_WORDS[@]} || lst=(${lst[@]} --only-installed)
		_kcpContains O only-outdated ${COMP_WORDS[@]} || lst=(${lst[@]} --only-outdated)
		_kcpContains f force-update ${COMP_WORDS[@]} || lst=(${lst[@]} --force-update)
	fi

	lst="${lst[@]}"
	[[ $lst == "" ]] && return 0
	COMPREPLY=( $(compgen -W "$lst" -- "$cur") )
}

complete -F _kcp kcp
