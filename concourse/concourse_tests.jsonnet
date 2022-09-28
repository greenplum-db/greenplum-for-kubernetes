local concourse = import "concourse.libsonnet";

local job = {
    name: "test job",
    plan: [
        { task: "task 0" },
        { put: "not a task" },
        concourse.InParallel([
            { task: "task 2.0" },
            { task: "task 2.1" },
            concourse.Do([
                { task: "task 2.2.0" },
                { task: "task 2.2.1" },
            ]),
        ]),
        concourse.Do([
            { task: "task 3.0" },
            { task: "task 3.1" },
            { task: "task 3.2" },
        ]),
        { task: "task 4" },
    ],
};

[
    {
        description: "ModifyTask modifies top-level task and ignores non-task steps",
        local modified_job = job + concourse.ModifyTask(name="task 0", merge={modified: true}),
        expected: { task: "task 0", modified: true },
        actual: modified_job.plan[0],
        result: self.expected == self.actual,
    },
    {
        description: "ModifyTask modifies task nested in an in_parallel step",
        local modified_job = job + concourse.ModifyTask(name="task 2.1", merge={modified: true}),
        expected: { task: "task 2.1", modified: true },
        actual: modified_job.plan[2].in_parallel[1],
        result: self.expected == self.actual,
    },
    {
        description: "ModifyTask modifies task nested in a do step",
        local modified_job = job + concourse.ModifyTask(name="task 3.0", merge={modified: true}),
        expected: { task: "task 3.0", modified: true },
        actual: modified_job.plan[3].do[0],
        result: self.expected == self.actual,
    },
    {
        description: "ModifyTask recurses deeply",
        local modified_job = job + concourse.ModifyTask(name="task 2.2.1", merge={modified: true}),
        expected: { task: "task 2.2.1", modified: true },
        actual: modified_job.plan[2].in_parallel[2].do[1],
        result: self.expected == self.actual,
    },
] +

local pipeline = {
    jobs: [
        { name: "job 1", group: "A" },
        { name: "job 2", group: "C" },
        { name: "job 3", group: "B" },
        { name: "job 4", group: "A" },
    ]
};
[
    {
        description: "AutoGroups generates groups and an All group from jobs",
        local modified_pipeline = concourse.AutoGroups() + pipeline,
        expected: [
            { name: "A", jobs: ["job 1", "job 4"]},
            { name: "B", jobs: ["job 3"]},
            { name: "C", jobs: ["job 2"]},
            { name: "All", jobs: ["job 1", "job 2", "job 3", "job 4"]},
        ],
        actual: modified_pipeline.groups,
        result: self.expected == self.actual,
    },
    {
        description: "AutoGroups takes groupOrder to control the order of groups",
        local modified_pipeline = concourse.AutoGroups(groupOrder=["A", "C", "B"]) + pipeline,
        expected: [
            { name: "A", jobs: ["job 1", "job 4"]},
            { name: "C", jobs: ["job 2"]},
            { name: "B", jobs: ["job 3"]},
            { name: "All", jobs: ["job 1", "job 2", "job 3", "job 4"]},
        ],
        actual: modified_pipeline.groups,
        result: self.expected == self.actual,
    },
    {
        description: "AutoGroups(groupOrder) allows partial specification",
        local modified_pipeline = concourse.AutoGroups(groupOrder=["C", "B"]) + pipeline,
        expected: [
            { name: "C", jobs: ["job 2"]},
            { name: "B", jobs: ["job 3"]},
            { name: "A", jobs: ["job 1", "job 4"]},
            { name: "All", jobs: ["job 1", "job 2", "job 3", "job 4"]},
        ],
        actual: modified_pipeline.groups,
        result: self.expected == self.actual,
    },
] +

local pipeline = {
    jobs: [
        { name: "artful aardvark", group: "A" },
        { name: "lucid lynx", group: "A" },
        { name: "xenial xerus", group: "A" },
        { name: "Breezy Badger", group: "A" },
    ]
};
[
    {
        description: "AutoGroups sorts jobs in a group so they remain in a stable order",
        local modified_pipeline = concourse.AutoGroups() + pipeline,
        expected: [
            { name: "A", jobs: ["Breezy Badger", "artful aardvark", "lucid lynx", "xenial xerus"]},
            { name: "All", jobs: ["Breezy Badger", "artful aardvark", "lucid lynx", "xenial xerus"]},
        ],
        actual: modified_pipeline.groups,
        result: self.expected == self.actual,
    },
] +

local topLevelObjectsTestPipeline = {
    local this = self,
    prod:: error 'required: prod',
    tla:: { prod: this.prod },
    resources: [
        {name: "prod-and-dev", dev_pipeline:: "yes"},
        {name: "prod-only", dev_pipeline:: "no"},
        {name: "dev-only", dev_pipeline:: "only"},
        {name: "prod-only-2"},
    ],
    jobs: [
        {name: "prod-and-dev", dev_pipeline:: "yes"},
        {name: "prod-only", dev_pipeline:: "no"},
        {name: "dev-only", dev_pipeline:: "only"},
        {name: "prod-only-2"},
    ],
};
local jobOnFailureTestPipeline = {
    local this = self,
    prod:: error 'required: prod',
    tla:: { prod: this.prod },
    resources: [],
    jobs: [
        {name: "prod-and-dev", dev_pipeline:: "yes",
            on_failure: {put: "slack-alert", dev_pipeline:: "yes"},
        },
        {name: "prod-only", dev_pipeline:: "yes",
            on_failure: {put: "slack-alert", dev_pipeline:: "no"},
        },
        {name: "dev-only", dev_pipeline:: "yes",
            on_failure: {put: "slack-alert", dev_pipeline:: "only"},
        },
        {name: "prod-and-dev-2", dev_pipeline:: "yes",
            on_failure: {put: "slack-alert"},
        },
        {name: "prod-and-dev-do", dev_pipeline:: "yes",
            on_failure: {do: [{put: "slack-alert" }]},
        },
        {name: "prod-only-do", dev_pipeline:: "yes",
            on_failure: {do: [{put: "slack-alert", dev_pipeline:: "no"}]},
        },
        {name: "dev-only-do", dev_pipeline:: "yes",
            on_failure: {do: [{put: "slack-alert", dev_pipeline:: "only"}]},
        },
        {name: "prod-and-dev-do-2", dev_pipeline:: "yes",
            on_failure: {do: [{put: "slack-alert", dev_pipeline:: "yes"}]},
        },
    ],
};
[
    {
        description: "FilterPipelineForProdOrDev selects top-level objects with dev_pipeline='yes' or 'only' for a dev pipeline",
        expected: {
            resources: [
                {name: "prod-and-dev"},
                {name: "dev-only"},
            ],
            jobs: [
                {name: "prod-and-dev"},
                {name: "dev-only"},
            ],
        },
        actual: topLevelObjectsTestPipeline{prod: false} + concourse.FilterPipelineForProdOrDev,
        result: self.expected == self.actual,
    },
    {
        description: "FilterPipelineForProdOrDev selects top-level objects with dev_pipeline='yes', 'no', or unmarked for a prod pipeline",
        expected: {
            resources: [
                {name: "prod-and-dev"},
                {name: "prod-only"},
                {name: "prod-only-2"},
            ],
            jobs: [
                {name: "prod-and-dev"},
                {name: "prod-only"},
                {name: "prod-only-2"},
            ],
        },
        actual: topLevelObjectsTestPipeline{prod: true} + concourse.FilterPipelineForProdOrDev,
        result: self.expected == self.actual,
    },
    {
        description: "FilterPipelineForProdOrDev filters out top-level objects with a bogus dev_pipeline value for a dev pipeline",
        local pipeline = {
            tla:: { prod: false },
            resources: [
                {name: "junk", dev_pipeline:: "junk"},
            ],
            jobs: [],
        },
        expected: { resources: [], jobs: [] },
        actual: pipeline + concourse.FilterPipelineForProdOrDev,
        result: self.expected == self.actual,
    },
    {
        description: "FilterPipelineForProdOrDev selects job.on_failure steps with dev_pipeline='yes', 'only', or unmarked for a dev pipeline",
        expected: {
            resources: [],
            jobs: [
                {name: "prod-and-dev", on_failure: {put: "slack-alert"}},
                {name: "prod-only"},
                {name: "dev-only", on_failure: {put: "slack-alert"}},
                {name: "prod-and-dev-2", on_failure: {put: "slack-alert"}},
                {name: "prod-and-dev-do", on_failure: {do: [{put: "slack-alert"}]}},
                {name: "prod-only-do", on_failure: {do: []}},
                {name: "dev-only-do", on_failure: {do: [{put: "slack-alert"}]}},
                {name: "prod-and-dev-do-2", on_failure: {do: [{put: "slack-alert"}]}},
            ],
        },
        actual: jobOnFailureTestPipeline{prod: false} + concourse.FilterPipelineForProdOrDev,
        result: self.expected == self.actual,
    },
    {
        description: "FilterPipelineForProdOrDev selects job.on_failure steps with dev_pipeline='yes', 'no', or unmarked for a prod pipeline",
        expected: {
            resources: [],
            jobs: [
                {name: "prod-and-dev", on_failure: {put: "slack-alert"}},
                {name: "prod-only", on_failure: {put: "slack-alert"}},
                {name: "dev-only"},
                {name: "prod-and-dev-2", on_failure: {put: "slack-alert"}},
                {name: "prod-and-dev-do", on_failure: {do: [{put: "slack-alert"}]}},
                {name: "prod-only-do", on_failure: {do: [{put: "slack-alert"}]}},
                {name: "dev-only-do", on_failure: {do: []}},
                {name: "prod-and-dev-do-2", on_failure: {do: [{put: "slack-alert"}]}},
            ],
        },
        actual: jobOnFailureTestPipeline{prod: true} + concourse.FilterPipelineForProdOrDev,
        result: self.expected == self.actual,
    },
    {
        description: "FilterPipelineForProdOrDev filters out job.on_failure steps with a bogus dev_pipeline value for a dev pipeline",
        local pipeline = {
            tla:: { prod: false },
            resources: [],
            jobs: [
                {name: "junk", dev_pipeline:: "yes",
                    on_failure: { put: "slack-alert", dev_pipeline:: "junk"},
                },
                {name: "junk-do", dev_pipeline:: "yes",
                    on_failure: { do: [{put: "slack-alert", dev_pipeline:: "junk"}]},
                },
            ],
        },
        expected: {
            resources: [],
            jobs: [
                {name: "junk"},
                {name: "junk-do", on_failure: { do: [] }},
            ],
        },
        actual: pipeline + concourse.FilterPipelineForProdOrDev,
        result: self.expected == self.actual,
    },
]
