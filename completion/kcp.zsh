#compdef kcp

#typeset -A opt_args
#setopt extendedglob

_kcpApps() {
	for a in $(kcp -lN | sort); do
		compadd $a
	done
}

_kcp() {
	_arguments -s : \
		- '(help)' \
			{-h,--help}"[Display usage]" \
		- '(version)' \
			{-v,--version}"[Display version]" \
		- '(update)' \
			{-u,--update-database}"[Update database]" \
		- list \
			'(l list)'{-l,--list}"[List packages]" \
			'(x sort)'{-x,--sort}"[Sort by popularity]" \
			'(f force)'{-f,--force-update}"[Force update]" \
			'(n name)'{-N,--only-name}"[Display only names]" \
			'(S starred)'{-S,--only-starred}"[Display only popular packages]" \
			'(I installed)'{-I,--only-installed}"[Display only installed packages]" \
			'(O outdated)'{-O,--only-outdated}"[Display only outdated packages]" \
		- search \
			'(s search)'{-s,--search}"[Search package]:package:_kcpApps" \
			{-x,--sort}"[Sort by popularity]" \
			{-f,--force-update}"[Force update]" \
			{-N,--only-name}"[Display only names]" \
			{-S,--only-starred}"[Display only popular packages]" \
			{-I,--only-installed}"[Display only installed packages]" \
			{-O,--only-outdated}"[Display only outdated packages]" \
		- install \
			'(i install)'{-i,--install}"[Install package]:package:_kcpApps" \
			'(D asdeps)'{-D,--asdeps}"[Install as a depend]" \
		- '(get)' \
			{-g,--get}"[Get needed files to build the package]:package:_kcpApps" \
		- '(information)' \
			{-v,--information}"[Get details about a package]:package:_kcpApps"
}

_kcp "$@"