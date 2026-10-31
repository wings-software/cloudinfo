#!/bin/sh
# Copyright 2025 Harness Inc. All rights reserved.
# Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
# that can be found in the licenses directory at the root of this repository, also available at
# https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

# Entrypoint for the harness-beta FIPS images (openjdk-fips, python-fips,
# go-fips). It changes nothing unless FIPS_ENABLED=false, in which case FIPS is
# switched off for the command it runs. It never switches FIPS on: without the
# flag the image runs exactly as shipped.
#
# POSIX sh, not bash: openjdk-fips and go-fips ship only busybox sh.
#
# Only settings the image itself declares are removed; no FIPS path or flag
# value is hard-coded here, so this follows the image as it changes.

# remove_words LIST REMOVE SEP: prints LIST without the SEP-separated words in
# REMOVE. Globbing is off so classpath entries like ./* stay literal.
remove_words() {
  _out=
  _ifs=$IFS
  IFS=$3
  set -f
  for _w in $1; do
    case "$3$2$3" in *"$3$_w$3"*) continue ;; esac
    _out=${_out:+$_out$3}$_w
  done
  set +f
  IFS=$_ifs
  printf '%s' "$_out"
}

if [ "$FIPS_ENABLED" = "false" ]; then
  # OpenSSL: /etc/ssl/openssl.cnf loads only the FIPS provider. The non-FIPS
  # images ship no openssl.cnf; an empty config gives the same built-in defaults.
  export OPENSSL_CONF=/dev/null

  # Java: the image keeps its FIPS JVM flags and classpath in
  # JDK_JAVA_FIPS_OPTIONS and JAVA_FIPS_CLASSPATH and copies them into
  # JDK_JAVA_OPTIONS and CLASSPATH. Strip just those, keeping anything the
  # deployment added. Unset when empty: an empty JDK_JAVA_OPTIONS still makes
  # the JVM print "NOTE: Picked up JDK_JAVA_OPTIONS:".
  if [ -n "$JDK_JAVA_FIPS_OPTIONS" ]; then
    opts=$(remove_words "$JDK_JAVA_OPTIONS" "$JDK_JAVA_FIPS_OPTIONS $JAVA_TRUSTSTORE_OPTIONS" " ")
    if [ -n "$opts" ]; then export JDK_JAVA_OPTIONS="$opts"; else unset JDK_JAVA_OPTIONS; fi
  fi
  if [ -n "$JAVA_FIPS_CLASSPATH" ]; then
    cp=$(remove_words "$CLASSPATH" "$JAVA_FIPS_CLASSPATH" ":")
    if [ -n "$cp" ]; then export CLASSPATH="$cp"; else unset CLASSPATH; fi
  fi
  unset JDK_JAVA_FIPS_OPTIONS JAVA_FIPS_CLASSPATH JAVA_TRUSTSTORE_OPTIONS

  # Go: binaries built with GOFIPS140 default to fips140=on, and GODEBUG
  # overrides that at run time. Replace any fips140 setting, keep the rest.
  godebug=
  _ifs=$IFS
  IFS=,
  set -f
  for kv in $GODEBUG; do
    case "$kv" in fips140=*) ;; *) godebug=${godebug:+$godebug,}$kv ;; esac
  done
  set +f
  IFS=$_ifs
  export GODEBUG="${godebug:+$godebug,}fips140=off"
  # go-fips: keep binaries built in this container from defaulting to FIPS.
  if [ -n "$GOFIPS140" ]; then export GOFIPS140=off; fi

  echo "FIPS mode disabled" >&2
fi

if [ $# -eq 0 ]; then
  echo "ERROR: no command given; set CMD in the Dockerfile (ENTRYPOINT resets the base image's CMD)" >&2
  exit 1
fi

# Run the actual app passed as CMD
exec "$@"
