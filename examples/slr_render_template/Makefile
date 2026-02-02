all:
	make render-single-template
	make render-multiple-templates

render-single-template:
	slr render-template \
		--template src/templates/README.md.gotemplate \
		--output output/single/README.generated.md \
		--data project=./src/data/project.yaml \
		--data users=./src/data/users.yaml

	cat output/single/README.generated.md

render-multiple-templates:
	slr render-templates

	cat output/multiple/README.generated.md
	cat output/multiple/CONTACTS.generated.md
