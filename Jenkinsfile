node('centos7') {
	def name = "netmgmt"
	def orgName = "github.com/swisstxt"
	def stageFilter = /(?:release|hotfix)-([0-9]+(?:\.[0-9]+))*/
	
	def workspaceDir = env.WORKSPACE
	def specsDir = "${workspaceDir}/SPECS"
	def sourcesDir = "${workspaceDir}/SOURCES"
	def relProjectOrgDir = "SOURCES/src/${orgName}"
	def projectOrgDir = "${workspaceDir}/${relProjectOrgDir}"
	def relProjectSourceDir = "${relProjectOrgDir}/${name}"
	def projectSourceDir = "${workspaceDir}/${relProjectSourceDir}"
	def relRpmbuildDir = "rpmbuild"
	def rpmbuildDir = "${workspaceDir}/${relRpmbuildDir}"
	
	def branch = ''
	def version = ''
	def release = env.BUILD_NUMBER
	def spec = ''
	def arch = ''
	def osRelease = ''
	def rev = ''
	// declaring this explicitly does not work thanks to the idiotic Jenkins sandbox
	//def stage = false
	
	stage('Checkout Repo') {
		checkout scm
		echo env.GIT_COMMIT
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
		branch = sh(
			script: "git rev-parse --abbrev-ref HEAD",
			returnStdout: true
		).trim()
		
		// necessary due to sandbox limitations on closures
		def lStage
		def lVersion
		
		// the current branch is just 'HEAD' if no explicit branch was checked out
		if (branch == 'HEAD') {
			lStage = ''
			lVersion = sh(
				script: "/opt/buildhelper/buildhelper getgittag ${workspaceDir}",
				returnStdout: true
			).trim()
		} else {
			def branchMatch = (branch =~ stageFilter)
			if (branchMatch) {
				lStage = 'stage-';
				lVersion = branchMatch[0][1]
			} else {
				error "Cannot determine version to build reliably. Exiting."
			}
		}

		stage = lStage
		version = lVersion
		release = "${release}.${rev}"
		env.GOPATH = sourcesDir
		env.PATH = "${sourcesDir}/bin:${env.PATH}"
		
		echo "name=${name}"
		echo "branch=${branch}"
		echo "version=${version}"
		echo "release=${release}"
	}
	
	stage('Prepare Build') {
		sh "mkdir -p ${projectOrgDir}"
		sh "ln -sf ${workspaceDir} ${projectSourceDir}"
	}
	
	stage('Clean Before Build') {
		for (path in [rpmbuildDir, "${sourcesDir}/pkg", "${sourcesDir}/lib"]) {
			sh "rm -rf ${path}"
		}
		sh "rm -f ${sourcesDir}/netmgmt.bin"
		for (path in ["${rpmbuildDir}/SPECS", "${rpmbuildDir}/SOURCES", specsDir, "${sourcesDir}/pkg", "${sourcesDir}/lib"]) {
			sh "mkdir -p ${path}"
		}
	}
	
	stage('Get Dependencies') {
		sh "go get github.com/tools/godep"
	}
	
	stage('Compile Source') {
		dir(projectSourceDir) {
			sh """
				cd ${projectSourceDir}
				pwd
				godep restore
				godep go install
			"""
		}
		sh "cp ${sourcesDir}/bin/netmgmt ${sourcesDir}/netmgmt.bin"
	}
	
	stage('Build RPM') {
		sh "cp -r ${specsDir}/* ${rpmbuildDir}/SPECS/ || true"
		sh "cp -r ${sourcesDir}/* ${rpmbuildDir}/SOURCES/ || true"
		sh """
			rpmbuild -ba ${spec} \
			--define "ver ${version}" \
			--define "rel ${release}" \
			--define "name ${name}" \
			--define "os_rel ${osRelease}" \
			--define "arch ${arch}" \
			--define "_topdir ${rpmbuildDir}" \
			--define "_builddir %{_topdir}" \
			--define "_rpmdir %{_topdir}" \
			--define "_srcrpmdir %{_topdir}"
		"""
	}
	
	stage('Archive RPM') {
		archiveArtifacts "${relRpmbuildDir}/*.rpm"
		archiveArtifacts "${relRpmbuildDir}/*/*.rpm"
	}
	
	stage('Push RPM') {
		echo "Would execute: /opt/buildhelper/buildhelper pushrpm yum-01.stxt.media.int:8080/swisstxt-centos"
	}
	
	stage('Deploy RPM') {
		if (stage != '') {
			build job: 'deploy.install.genericstxt', parameters: [
				string(name: 'INVENTORY_HOST', value: 'pcache'),
				string(name: 'INVENTORY_NAME', value: 'hosts/integration'),
				string(name: 'PLAY_TYPE', value: 'task'),
				string(name: 'PLAY_MODULE', value: 'shell'),
				text(name: 'PLAY_ARGUMENTS', value: 'echo "Executing: yum clean metadata && yum update netmgmt"'),
			]
		} else {
			echo "Release builds are not auto-deployed."
		}
	}
}
