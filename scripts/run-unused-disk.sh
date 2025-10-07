#!/usr/bin/env bash

set -euo pipefail

# Path to ksonnet/lib/meta/raw/clusters.json
CLUSTERS="$1"
EXT=$2

startup_asserts() {
	if [[ "$EXT" != "csv" && "$EXT" != "txt" ]]; then
		echo "Error: EXT must be 'csv' or 'txt'."
		exit 1
	fi
}

post_clean() {
	# Cleaning up some of the CSV mess
	# TODO: Make this not necessary
	if [[ "$EXT" == "csv" ]]; then
		echo "Cleaning up"
		rm $(grep -l "No disks found" ./data/*.csv)
		sed -i '' 's/:/_/g' ./data/*.csv 
		echo
	fi
}

main() {
	startup_asserts
	
	(
		< $CLUSTERS jq -r '.clusters[] | select(.project) | @text "\(.provider) \(.project)"'
		< $CLUSTERS jq -r '.clusters[] | select(.provider == "aks") | @text "aks \(.subscription_id)"'
	) |
		sort -u |
		grep -v govcloud |
		while read provider id; do
			case "$provider" in
				aks)
						flag="-azure.sub"
						;;
				eks)
						flag="-aws.profile"
						;;
				gke)
						flag="-gcp.project"
						;;
				*)
						echo "Unrecognized provider $provider"
						exit 1
						;;
			esac

			file="${provider}.${id}.${EXT}"

			echo "$provider $id"
			if [[ "$EXT" == "csv" ]]; then
				./unused "${flag}=${id}" -add-k8s-column=ns -add-k8s-column=pvc -add-k8s-column=pv -v -csv > ./data/$file
			else
				./unused "${flag}=${id}" -add-k8s-column=ns -add-k8s-column=pvc -add-k8s-column=pv -v > ./data/$file
			fi
		done
}

main
