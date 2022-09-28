{
// Allows using "this.tla" in any part of the definition of "this" so that this
// file is the only place we have to specify what the top-level arguments are.
// Returns a function (appropriate to be used at the top level) to inject the
// arguments as a hidden field of `this` named `tla`
// Example usage:
/*
$ cat test.jsonnet
local tla = import "tla.libsonnet";

local minipipeline = {
    local this = self,
    jobs+: [ { name: this.tla.pipeline_name } ],
};

local main_pipeline = {
    local this = self,
    jobs:[],
    resources:[{branch: this.tla.git_branch}],
};

tla.Apply(main_pipeline + minipipeline)

$ jsonnet test.jsonnet --tla-str pipeline_name=foo
{
   "jobs": [
      {
         "name": "foo"
      }
   ],
   "resources": [
      {
         "branch": "master"
      }
   ]
}
*/
Apply(this):: function(prod=true, pipeline_name="", git_branch="master")
    this {
        tla:: {
            prod: prod,
            pipeline_name: pipeline_name,
            git_branch: git_branch,
        }
    }
}
