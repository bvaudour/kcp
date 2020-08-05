#compdef kcp

#typeset -A opt_args
#setopt extendedglob

_kcp_help=( '(-h,--help)'{-h,--help}'[Display usage]' )
_kcp_version=( '(-v,--version)'{-v,--version}'[Display version]' )
_kcp_update=( '(-u,--update-database)'{-u,--update-database}'[Update database]' )
_kcp_get=( '(-g,--get)'{-g,--get}'[Get needed files to build the package]:package:_kcpApps' )
_kcp_information=( '(-V,--information)'{-V,--information}'[Get details about a package]:package:_kcpApps' )
_kcp_list=( '(-l,--list)'{-l,--list}'[List packages]' )
_kcp_search=( '(-s,--search)'{-s,--search}'[Search package]:package:_kcpApps' )
_kcp_install=( '(-i,--install)'{-i,--install}'[Install package]:package:_kcpApps' )
_kcp_sort=( '(-x,--sort)'{-x,--sort}'[Sort by popularity]' )
_kcp_force=( '(-f,--force-update)'{-f,--force-update}'[Force update]' )
_kcp_name=( '(-N,--only-name)'{-N,--only-name}'[Display only names]' )
_kcp_starred=( '(-S,--only-starred)'{-S,--only-starred}'[Display only popular packages]' )
_kcp_installed=( '(-I,--only-installed)'{-I,--only-installed}'[Display only installed packages]' )
_kcp_outdated=( '(-O,--only-outdated)'{-O,--only-outdated}'[Display only outdated packages]' )
_kcp_asdeps=( '(-D,--asdeps)'{-D,--asdeps}'[Install as a depend]' )

_kcpApps() {
	for a in $(kcp -lN | sort); do
		compadd $a
	done
}

_kcp() {
	_arguments -s : \
		- '(help)' \
			"$_kcp_help[@]" \
		- '(version)' \
			"$_kcp_version[@]" \
		- '(update)' \
			"$_kcp_update[@]" \
		- '(get)' \
			"$_kcp_get[@]" \
		- '(information)' \
			"$_kcp_information[@]" \
		- list \
			"$_kcp_list[@]" \
			"$_kcp_sort[@]" \
			"$_kcp_force[@]" \
			"$_kcp_name[@]" \
			"$_kcp_starred[@]" \
			"$_kcp_installed[@]" \
			"$_kcp_outdated[@]" \
		- search \
			"$_kcp_search[@]" \
			"$_kcp_sort[@]" \
			"$_kcp_force[@]" \
			"$_kcp_name[@]" \
			"$_kcp_starred[@]" \
			"$_kcp_installed[@]" \
			"$_kcp_outdated[@]" \
		- install \
			"$_kcp_install[@]" \
			"$_kcp_asdeps[@]"
}

_kcp "$@"
