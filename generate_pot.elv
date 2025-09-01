#!/usr/bin/env elvish

var workdir  = cache
var alldir   = all
var transdir = kcp

if ?(test -e $workdir) {
  rm -rf $workdir
}
mkdir $workdir

var ignored = "^(vendor|resources|cache|build|"$workdir")"

fn create_base {||
  each {|l| echo $l >> $workdir/base.pot } [
    '#, fuzzy'
    'msgid ""'
    'msgstr ""'
    '"Project-Id-Version: PACKAGE VERSION\n"'
    '"Report-Msgid-Bugs-To: \n"'
    '"POT-Creation-Date: 2020-08-06 12:01+0200\n"'
    (printf '"PO-Revision-Date: %s\n"'  (date +'%F %R%z'))
    '"Last-Translator: FULL NAME <EMAIL@ADDRESS>\n"'
    '"Language-Team: LANGUAGE <LL@li.org>\n"'
    '"Language: \n"'
    '"MIME-Version: 1.0\n"'
    '"Content-Type: text/plain; charset=UTF-8\n"'
    '"Content-Transfer-Encoding: 8bit\n"'
    ''
  ]
}

fn extract_all {||
  i18n4go -c extract-strings -v --po -d . -r -o $workdir/$alldir --ignore-regexp $ignored -output-match-package
}

fn extract_trans {||
  put **/translatable.go | each {|file|
    var d = $workdir/$transdir/(dirname $file)
    i18n4go -c extract-strings -v --po -f $file -o $d --ignore-regexp $ignored
  }
}

fn merge_json {|d|
  i18n4go -c merge-strings -d $workdir/$d -r -source-language en
  var json = []
  put $workdir/$d/**/all.en.json | each {|f|
    var j = (cat $f | from-json)
    set @j = (all $j | each {|e| dissoc $e modified })
    set @json = $@json $@j
  }
  put $json | to-json > $workdir/$d/kcp.json
}

fn convert_json {|d|
  json2po -P --filter=id --duplicates=merge $workdir/$d/kcp.{json,pot}
}

fn merge_po {|d|
  put $workdir/$d/**.po | each {|f|
    e:cat $workdir/base.pot $f > $f.tmp
    e:mv $f{.tmp,}
  }
  #xgettext -o $workdir/$d/kcp.pot $workdir/$d/**.po --from-code=UTF-8 --no-wrap --extract-all
  msgcat -u -o $workdir/$d/kcp.po $workdir/$d/**.po
  msgattrib -o $workdir/$d/kcp.{pot,po}
}

create_base

if (> (count $args) 0) {
  var e = $args[0]
  if (is $e --all) {
    extract_all
    merge_json $alldir
    #convert_json $alldir
    merge_po $alldir
  }
}

extract_trans
merge_po $transdir
