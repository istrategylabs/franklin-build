tmp_id=$RANDOM
project_name=$1
docker build -t tmp_id .
docker run -v `pwd`/build_directory:/tmp_mount tmp_id cp -r /$project_name/dist /tmp_mount
