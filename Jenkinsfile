node('centos7') {
	// In most cases, you only need to change the sections marked as "ADAPT"
	
	// ADAPT
	// Name of the project and package
	def name = 'netmgmt'
	// Qualified organization name (for Go projects)
	def orgName = 'github.com/swisstxt'
	// Regex for identifying stage branch names
	def stageFilter = /(?:release|hotfix)-([0-9]+(?:\.[0-9]+))*/
	// Suffix for staging packages
	def stageSuffix = '-stage'
	// END ADAPT

	def workspaceDir = env.WORKSPACE
	def specsDir = "${workspaceDir}/SPECS"
	def sourcesDir = "${workspaceDir}/SOURCES"
	def relProjectOrgDir = "SOURCES/src/${orgName}"
	def projectOrgDir = "${workspaceDir}/${relProjectOrgDir}"
	def relProjectSourceDir = "${relProjectOrgDir}/${name}"
	def projectSourceDir = "${workspaceDir}/${relProjectSourceDir}"
	def relRpmbuildDir = "rpmbuild"
	def rpmbuildDir = "${workspaceDir}/${relRpmbuildDir}"
	
	def buildNumber = env.BUILD_NUMBER
	def branch = ''
	def version = ''
	def release = ''
	def spec = ''
	def arch = ''
	def osRelease = ''
	def rev = ''
	def isStaging = false
	def pkgName = name
	
	stage('Checkout Repo') {
		checkout scm
	}
		
	stage('Set Build Variables') {
		spec = sh(
			script: "/opt/buildhelper/buildhelper getspec ${name}",
			returnStdout: true
		).trim()
		arch = sh(
			script: "/opt/buildhelper/buildhelper getarch || true",
			returnStdout: true
		).trim()
		osRelease = sh(
			script: "/opt/buildhelper/buildhelper getosrelease || true",
			returnStdout: true
		).trim()
		rev = sh(
			script: "git rev-parse --short HEAD",
			returnStdout: true
		).trim()
		// Should be GIT_BRANCH, but does not work due to #JENKINS-35230 and #SECURITY-170
		// Needs "Check out to local branch: **" in Jenkins,
		// but do not set for tag builds.
		branch = sh(
			script: "git rev-parse --abbrev-ref HEAD",
			returnStdout: true
		).trim()
		
		// the current branch is just 'HEAD' if no explicit branch was checked out
		if (branch == 'HEAD') {
			isStaging = false
		} else {
			def branchMatch = (branch =~ stageFilter)
			if (branchMatch) {
				isStaging = true
				version = branchMatch[0][1]
			} else {
				isStaging = false
			}
		}
		if (isStaging) {
			pkgName = "${name}${stageSuffix}"
			release = "${buildNumber}.${rev}"
		} else {
			pkgName = "${name}"
			release = "${buildNumber}"
			version = sh(
				script: "/opt/buildhelper/buildhelper getgittag ${workspaceDir}",
				returnStdout: true
			).trim()
			// TODO remove the tag prefix (i.e. 'v')
		}
		
		env.GOPATH = sourcesDir
		env.PATH = "${sourcesDir}/bin:${env.PATH}"
		
		echo "pkgName=${pkgName}"
		echo "branch=${branch}"
		echo "version=${version}"
		echo "release=${release}"
	}
	
	stage('Prepare Build') {
		// Set up Go-friendly source tree
		sh "mkdir -p ${projectOrgDir}"
		// and link source code there
		sh "ln -sf ${workspaceDir} ${projectSourceDir}"
	}
	
	stage('Clean Before Build') {
		// Remove built files directories
		for (path in [rpmbuildDir, "${sourcesDir}/pkg", "${sourcesDir}/lib"]) {
			sh "rm -rf ${path}"
		}
		sh "rm -f ${sourcesDir}/netmgmt.bin"
		for (path in ["${rpmbuildDir}/SPECS", "${rpmbuildDir}/SOURCES", specsDir, "${sourcesDir}/pkg", "${sourcesDir}/lib"]) {
			sh "mkdir -p ${path}"
		}
	}
	
	stage('Get Dependencies') {
		// Fetch godep tool
		sh "go get github.com/tools/godep"
	}
	
	stage('Compile Source') {
		// ADAPT
		// Fetch dependencies and build source using godep
		dir(projectSourceDir) {
			sh """
				cd ${projectSourceDir}
				pwd
				godep restore
				godep go install
			"""
		}
		// Copy the resulting binary
		sh "cp ${sourcesDir}/bin/netmgmt ${sourcesDir}/netmgmt.bin"
		// END ADAPT
	}
	
	stage('Build RPM') {
		// Copy the results into the RPM build environment
		sh "cp -r ${specsDir}/* ${rpmbuildDir}/SPECS/ || true"
		sh "cp -r ${sourcesDir}/* ${rpmbuildDir}/SOURCES/ || true"
		// And build the RPM
		sh """
			rpmbuild -ba ${spec} \
			--define "ver ${version}" \
			--define "rel ${release}" \
			--define "name ${pkgName}" \
			--define "os_rel ${osRelease}" \
			--define "arch ${arch}" \
			--define "_topdir ${rpmbuildDir}" \
			--define "_builddir %{_topdir}" \
			--define "_rpmdir %{_topdir}" \
			--define "_srcrpmdir %{_topdir}"
		"""
	}
	
	stage('Archive RPM') {
		// Send the results to Jenkins
		archiveArtifacts "${relRpmbuildDir}/*.rpm"
		archiveArtifacts "${relRpmbuildDir}/*/*.rpm"
	}
	
	stage('Push RPM') {
		// Send the RPM to the package server
		// TODO Change to sh "/opt..."
		echo "Would execute: /opt/buildhelper/buildhelper pushrpm yum-01.stxt.media.int:8080/swisstxt-centos"
	}
	
	stage('Deploy RPM') {
		// ADAPT
		// To change these to manual, just comment out the respective build job line
		if (isStaging) {
			// Call the deploy job for staging
			build job: 'deploy.install.netmgmt', parameters: [ string(name: 'DEPLOY_ENVIRONMENT', value: 'stage') ]
		} else {
			// Call the deploy job for production
			build job: 'deploy.install.netmgmt', parameters: [ string(name: 'DEPLOY_ENVIRONMENT', value: 'production') ]
		}
		// END ADAPT
	}
}
