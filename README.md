milo
====

A go based web framework.

#  There are a few goals to the milo project.
1. Use as much of the default net/http package as possible
	- Register static routes for assets
2. Work with a specific go http design pattern
3. Routing Helpers
	- Provide a wrapper for routing to plug pre & post middleware into
	- Helpers for handling 404
	- Helpers for handling errors
4. Response Rendering
	- Rendering multiple templates
	- Rendering JSON output
	- Rendering error output