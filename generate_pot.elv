#!/usr/bin/env elvish

local:workdir  = cache
local:alldir   = all
local:transdir = kcp

if ?(test -e $workdir) {
  rm -rf $workdir
}
mkdir $workdir

local:ignored = "^(vendor|resources|cache|build|"$workdir")"

fn extract_all []{
  i18n4go -c extract-strings -v --po -d . -r -o $workdir/$alldir --ignore-regexp $ignored -output-match-package
}

fn extract_trans []{
  put **/consts.go | each [file]{
    local:d = $workdir/$transdir/(dirname $file)
    i18n4go -c extract-strings -v --po -f $file -o $d --ignore-regexp $ignored
  }
}

fn merge_json [d]{
  i18n4go -c merge-strings -d $workdir/$d -r -source-language en
  local:json = []
  put $workdir/$d/**/en.all.json | each [f]{
    local:j = (cat $f | from-json)
    @j = (all $j | each [e]{ dissoc $e modified })
    @json = $@json $@j
  }
  put $json | to-json > $workdir/$d/kcp.json
}

fn convert_json [d]{
  json2po -P $workdir/$d/kcp.{json,pot}
}

fn merge_po [d]{
  xgettext -o $workdir/$d/kcp.pot $workdir/$d/**.po
}

if (> (count $args) 0) {
  local:e = $args[0]
  if (is $e --all) {
    extract_all
    merge_json $alldir
    merge_po $alldir
  }
}

extract_trans
merge_po $transdir
