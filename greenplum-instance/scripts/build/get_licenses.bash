#!/usr/bin/env bash

# Just run this bash script inside the Ubuntu container to get all the required license information:
# +------+---------+-------------+-----------+
# | name | version | description | copyright |
# +------+---------+-------------+-----------+

FILE_NAME=/tmp/licenses.csv

rm -f ${FILE_NAME}

for package in $(dpkg --list | tail -n +6 | awk '{ print $2 }' | awk -F ":" '{ print $1 }'); do
        dpkg --list $package | tail -n 1 | awk '{ printf "%s, %s, %s, \"", $2, $3, $4 }' >> ${FILE_NAME}
        cat /usr/share/doc/$package/copyright | tr '\n' '\r' | tr -d '"' >> ${FILE_NAME}
        echo "\"" >> ${FILE_NAME}
done

# remove duplicated CR, which caused google import issue
tr -s '\r' < ${FILE_NAME} > "$FILE_NAME.tmp"
cp "$FILE_NAME.tmp" ${FILE_NAME}
rm "$FILE_NAME.tmp"