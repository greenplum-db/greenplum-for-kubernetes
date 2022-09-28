---
title: Using MADlib for Analytics
---

This topic describes how to configure the MADlib open-source library for scalable in-database analytics in <%=vars.product_name %>.

## <a id="about"></a>About MADlib in <%=vars.product_name_long %>

Unlike with other <%=vars.product_name %> distributions, <%=vars.product_name_long %> automatically installs the MADlib software as part of the Greenplum Docker image. For example, after initializing a new Greenplum cluster in Kubernetes, you can see that MADlib is available as an installed Debian Package:

``` bash
$ kubectl exec -it master-0 -- bash -c "dpkg -s madlib"
```
``` bash
Package: madlib
Status: install ok installed
Priority: optional
Section: devel
Installed-Size: 59035
Maintainer: dev@madlib.apache.org
Architecture: amd64
Version: 1.17.0
Description: Apache MADlib is an Open-Source Library for Scalable in-Database Analytics
```

To begin using MADlib, you simply use the `madpack` utility to add MADlib functions to your database, as described in the next section.

## <a id="madpack"></a>Adding MADlib Functions

To install the MADlib functions to a database, use the `madpack` utility. For example:

``` bash
$ kubectl exec -it master-0 -- bash -c "source ./.bashrc; madpack -p greenplum install"
```
``` bash
madpack.py: INFO : Detected Greenplum DB version 6.8.0.
madpack.py: INFO : *** Installing MADlib ***
madpack.py: INFO : MADlib tools version    = 1.17.0 (/usr/local/madlib/Versions/1.17.0/bin/../madpack/madpack.py)
madpack.py: INFO : MADlib database version = None (host=localhost:5432, db=gpadmin, schema=madlib)
madpack.py: INFO : Testing PL/Python environment...
madpack.py: INFO : > Creating language PL/Python...
madpack.py: INFO : > PL/Python environment OK (version: 2.7.12)
madpack.py: INFO : > Preparing objects for the following modules:
madpack.py: INFO : > - array_ops
madpack.py: INFO : > - bayes
madpack.py: INFO : > - crf
...
madpack.py: INFO : Installing MADlib:
madpack.py: INFO : > Created madlib schema
madpack.py: INFO : > Created madlib.MigrationHistory table
madpack.py: INFO : > Wrote version info in MigrationHistory table
madpack.py: INFO : MADlib 1.17.0 installed successfully in madlib schema.
```

This installs MADlib functions into the default schema named `madlib`. Execute `madpack -h` or see the [Greenplum MADlib Extension for Analytics](http://gpdb.docs.pivotal.io/5120/ref_guide/extensions/madlib.html) documentation for <%=vars.product_name %> Database for more information about using `madpack`.

## <a id="moreinfo"></a>Getting More Information

For more information about using MADlib, see:

* [Apache MADlib Website](http://madlib.apache.org/)
* [Apache MADlib Documentation](http://madlib.apache.org/documentation.html)
* [Apache MADlib Wiki](https://cwiki.apache.org/confluence/display/MADLIB/)
