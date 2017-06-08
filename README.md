# JTed - Jenkins Templates for Terraform Editor
Terraform uses providers to manage the lifecycle of resources in the datacentre.
The [Terraform Jenkins Provider](https://github.com/dihedron/terraform-provider-jenkis) 
allows to manage Jenkins jobs as resources; the underlying API requires the client
to POST a complex XML file to create and update a job; the XML does not have a
definite schema because its structure and contents depend on the set of plugins
currently installed on the Jenkins server.
The Terraform Jenkins Provider solves this by relying on a job definition 
template (in Golang text templates format) and a set of parameters that are
applied onto that template right before submitting the POST to the server.
Thus, the job template can be created once for all projects of the same kind 
running on the same server, and each project team can tweak its chaaracteristics 
by modifying the relevant parameters.
This utility makes it easier to create a job template from an existing job
configuration file (```config.xml```) by identifying value that can be mapped to
parameters and by generating the template and the set of parameters in HCL 
format, ready for use in a Terraform recipe.
