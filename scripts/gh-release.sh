#!/bin/bash
#
########################################################################
# This is the script used to publish a release on the Github repository.
# It does the following steps:
#
# - make a new release tag on GitHub based on the given version number,
# - build RPM packages, and
# - upload RPMs as assessts of the release tag.
########################################################################

function get_script_dir() {
    ## resolve the base directory of this executable
    local SOURCE=$1
    while [ -h "$SOURCE" ]; do
        # resolve $SOURCE until the file is no longer a symlink
        DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
        SOURCE="$(readlink "$SOURCE")"

        # if $SOURCE was a relative symlink,
        # we need to resolve it relative to the path
        # where the symlink file was located

        [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
    done

    echo "$( cd -P "$( dirname "$SOURCE" )" && pwd )"
}

function new_release_post_data() {
    t=1
    p=2
    cat <<EOF
{
    "tag_name": "${tag}",
    "tag_commitish": "master",
    "name": "${tag}",
    "body": "Release ${tag}",
    "draft": false,
    "prerelease": $2
}
EOF
}

if [ $# -ne 2 ]; then
    echo "$0 <release_tag> <is_prerelease>"
    exit 1
fi

tag=$1
pre=$2
gh_token=""

RPM_BUILD_ROOT=$HOME/rpmbuild

GH_ORG="Donders-Institute"
GH_REPO_NAME="hpc-utility"

GH_API="https://api.github.com"
GH_REPO="$GH_API/repos/$GH_ORG/$GH_REPO_NAME"
GH_RELE="$GH_REPO/releases"
GH_TAG="$GH_REPO/releases/tags/$tag"
GH_REPO_ASSET_PREFIX="https://uploads.github.com/repos/$GH_ORG/$GH_REPO_NAME/releases"

# check if version tag already exists
response=$(curl -X GET $GH_TAG 2>/dev/null)
eval $(echo "$response" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
if [ "$id" ]; then
    read -p "release tag already exists: ${tag}, continue? y/[n]: " cnt
    if [ "${cnt,,}" != "y" ]; then
        exit 1
    fi
fi

# make sure the go command is available
which go > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "golang is required for building RPM."
    exit 1
fi

while [ "$gh_token" == "" ]; do
    read -s -p "github personal access token: " gh_token
done

# create a new tag with current master branch
# if the $id of the release is not available.
if [ ! "$id" ]; then
    response=$(curl -H "Authorization: token $gh_token" -X POST --data "$(new_release_post_data ${tag} ${pre})" $GH_RELE)
    eval $(echo "$response" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
    [ "$id" ] || { echo "release tag not created successfully: ${tag}"; exit 1; }
fi

# copy over id to rid (release id)
rid=$id

mydir=$( get_script_dir $0 )
path_spec=${mydir}/../build/rpm/centos7.spec

## replace the release version in
out=$( VERSION=${tag} rpmbuild --undefine=_disable_source_fetch -bb ${path_spec} )
if [ $? -ne 0 ]; then
    echo "rpm build failure"
    exit 1
fi

## parse the RPM build output to get paths of output RPMs
rpms=( $( echo "${out}" | egrep -o -e 'Wrote:.*\.rpm' | sed 's/Wrote: //g' ) )

## upload RPMs as release assets 
if [ ${#rpms[@]} -gt 0 ]; then
    upload="y"
    read -p "upload ${#rpms[@]} RPMs as release assets?, continue? [y]/n: " upload
    for rpm in ${rpms[@]}; do
        echo ${rpm}
        if [ "${upload,,}" == "y" ]; then
            fname=$( basename $rpm )
            # check if the asset with the same name already exists
            id=""
            eval $(echo "$response" | grep -C3 "name.:.\+${fname}" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
            if [ "$id" != "" ]; then
                # delete existing asset
                echo "deleting asset: ${id} ..."
                curl -H "Authorization: token $gh_token" -X DELETE "${GH_RELE}/assets/${id}"
            fi
            # post new asset
            echo "uploading ${rpm} ..."
            GH_ASSET="${GH_REPO_ASSET_PREFIX}/${rid}/assets?name=$(basename $rpm)"
            resp_upload=$( curl --data-binary @${rpm} \
                                -H "Content-Type: application/octet-stream" \
				-H "Authorization: token $gh_token" $GH_ASSET )
        fi
    done
fi
