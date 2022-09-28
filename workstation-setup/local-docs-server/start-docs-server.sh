#! /bin/bash

pushd /gp-kubernetes-docs-book && \
	bundle install && \
	sed -i 's#processes.detect ->{ NullProcess.new } { |process| process.applicable_to?(section) }#processes.detect(Proc.new { NullProcess.new }) { |process| process.applicable_to?(section) }#g' \
	/var/lib/gems/2.5.0/gems/bookbindery-10.1.14/lib/bookbinder/preprocessing/preprocessor.rb && \
	sed -i 's#detect ->{""} {|text| text}#detect(Proc.new {""}) {|text| text}#g' /var/lib/gems/2.5.0/gems/bookbindery-10.1.14/lib/bookbinder/code_example_reader.rb && \
	bundle exec bookbinder bind local && \
	popd

pushd /gp-kubernetes-docs-book/final_app && \
	sed -i "s#ruby '~> 2.3.0'#ruby '~> 2.5.1'#g" Gemfile
	bundle install && \
	rackup -o 0.0.0.0
