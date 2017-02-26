#!/bin/sh
set -e

HISTORYFILE=$1
if [ -z "$HISTORYFILE" ]; then
	echo "USAGE: $0 <historyfile>"
	exit
fi

FILTEREXPR="railway or railway:signal:speed_limit or railway:signal:speed_limit_distant or railway:signal:combined or railway:signal:distant"
SNAPSHOTDIR="snapshots/"

mkdir -p $SNAPSHOTDIR

MONTHS="01 02 03 04 05 06 07 08 09 10 11 12"
YEARS="2013 2014 2015 2016 2017"
STOP_AT="201702"

for YEAR in $YEARS; do
	for MONTH in $MONTHS; do
		if [ $YEAR$MONTH -gt $STOP_AT ]; then
			exit
		fi

		echo $YEAR-$MONTH
		# perform snapshot
		#osmium time-filter --progress $HISTORYFILE $YEAR-$MONTH-01T00:00:00Z -o $SNAPSHOTDIR/$YEAR-$MONTH.pbf
		# prefilter
		#osmium-filter-simple -v -w -e "$FILTEREXPR" -o $SNAPSHOTDIR/$YEAR-$MONTH-railway.pbf $SNAPSHOTDIR/$YEAR-$MONTH.pbf
		#rm $SNAPSHOTDIR/$YEAR-$MONTH.pbf
	done
done
