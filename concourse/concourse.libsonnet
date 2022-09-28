{
ModifyTask(name, merge={}): $.ModifyTaskInComposite('plan', name, merge),

ModifyTaskInComposite(compositeType, name, merge={}):: {
    [compositeType]: [
        if std.objectHas(x, 'in_parallel') then
            x + $.ModifyTaskInComposite('in_parallel', name, merge)
        else if std.objectHas(x, 'do') then
            x + $.ModifyTaskInComposite('do', name, merge)
        else if std.objectHas(x, 'task') && x.task == name then
            x + merge
        else
            x
        for x in super[compositeType]
    ]
},

// Use as a pipeline mixin to automatically generate groups from the "group"
// hidden property of the jobs in the pipeline. Groups will be output
// alphabetically by default, but the order can be overridden with groupOrder.
AutoGroups(groupOrder=[]): {
    local pipeline = self,

    local allGroups = std.set([j.group for j in pipeline.jobs]),
    local unorderedGroups = std.setDiff(allGroups, std.set(groupOrder)),
    local groupNames = groupOrder + unorderedGroups,

    groups: [
        {
            name: group,
            jobs: std.sort([j.name for j in pipeline.jobs if j.group == group]),
        }
        for group in groupNames
    ] + [{
        name: "All",
        jobs: std.sort([j.name for j in pipeline.jobs]),
    }],
},

// Save a level of indentation
InParallel(steps=[]): {
    in_parallel: steps,
},
Do(steps=[]): {
    do: steps,
},

// dev_pipeline: "yes"|"no"|"only"|<missing>
IsDevPipelineYes(object):: std.objectHasAll(object, 'dev_pipeline') && object.dev_pipeline == 'yes',
IsDevPipelineNo(object):: std.objectHasAll(object, 'dev_pipeline') && object.dev_pipeline == 'no',
IsDevPipelineOnly(object):: std.objectHasAll(object, 'dev_pipeline') && object.dev_pipeline == 'only',
IsDevPipelineUnspecified(object):: !std.objectHasAll(object, 'dev_pipeline'),

// Top-level object `dev_pipeline` meaning:
//   - 'yes'       = prod & dev
//   - 'no'        = prod
//   - 'only'      = dev
//   - unspecified = prod
// Inner object `dev_pipeline` meaning:
//   - 'yes'       = prod & dev
//   - 'no'        = prod
//   - 'only'      = dev
//   - unspecified = prod & dev
FilterPipelineForProdOrDev: {
    local filterTopLevelObjects =
        if self.tla.prod then
            function(object)
                $.IsDevPipelineYes(object) ||
                    $.IsDevPipelineNo(object) ||
                    $.IsDevPipelineUnspecified(object)
        else
            function(object)
                $.IsDevPipelineYes(object) ||
                    $.IsDevPipelineOnly(object)
    ,
    local filterInnerObjects =
        if self.tla.prod then
            function(object)
                $.IsDevPipelineYes(object) ||
                    $.IsDevPipelineNo(object) ||
                    $.IsDevPipelineUnspecified(object)
        else
            function(object)
                $.IsDevPipelineYes(object) ||
                    $.IsDevPipelineOnly(object) ||
                    $.IsDevPipelineUnspecified(object)
    ,
    local filterJobSteps(job) =
        // todo: at some point, we should consider making this recursive and support filtering job steps
        if std.objectHas(job, 'on_failure') then
            if std.objectHas(job['on_failure'], 'do') then
                job + {
                    'on_failure': {
                        'do': [
                            step
                            for step in job['on_failure']['do']
                            if filterInnerObjects(step)
                        ]
                    }
                }
            else if !filterInnerObjects(job['on_failure']) then
                job + { 'on_failure':: '' }
            else
                job
        else
            job
    ,
    jobs: [
        filterJobSteps(job)
        for job in super['jobs']
        if filterTopLevelObjects(job)
    ],
    resources: [
        resource
        for resource in super['resources']
        if filterTopLevelObjects(resource)
    ]
},

}
