#!/bin/sh

set -x

MYSQL_ROOT_PASSWORD="password"

MAVEN_REPO="http://repo1.maven.org/maven2"

JETTY_VERSION="9.2.6.v20141205"
MYSQL_CONNECTOR_VERSION="5.1.34"

CLOUDSTACK_WAR_URL="http://jenkins.buildacloud.org/view/4.3/job/cloudstack-4.3-maven-build-noredist/lastSuccessfulBuild/artifact/client/target/cloud-client-ui-4.3.2.war"
CLOUDSTACK_REPO_URL="https://github.com/apache/cloudstack/tags/4.3.2"



get_jar() {
	if [ -z "$(find "$HOME/.m2" -name "$2*.jar")" ]; then
		mvn -DrepoUrl="$MAVEN_REPO" -DgroupId="$1" -DartifactId="$2" -Dversion="$3" dependency:get >/dev/null 2>/dev/null
	fi
	if [ -n "$(find "$HOME/.m2" -name "$2*.jar")" ]; then
		find "$HOME/.m2" -name "$2*.jar" -print -quit
	else
		# Couldn't install jetty-runner
		return 1
	fi
}


jetty_runner_path=$(get_jar 'org.eclipse.jetty' 'jetty-runner' "$JETTY_VERSION")
mysql_connector_path=$(get_jar 'mysql' 'mysql-connector-java' "$MYSQL_CONNECTOR_VERSION")

mkdir -p database-imports/

if [ ! -f 'cloudstack.war' ]; then
	wget -O 'cloudstack.war' "$CLOUDSTACK_WAR_URL"
	jar -xvf cloudstack.war
	svn export --force "$CLOUDSTACK_REPO_URL/utils/conf/db.properties"
	svn export --force "$CLOUDSTACK_REPO_URL/developer/developer-prefill.sql" "database-imports/developer-prefill.sql"
	svn export --force "$CLOUDSTACK_REPO_URL/setup/db" "database-imports"
	rm -rf 'db'; mv 'database-imports/db' 'db'
fi

mkdir -p lib
cp "$mysql_connector_path" lib/

classpath=$(find . -name '*.jar' -print0 | sed -e 's/\.\//:/g')
java -cp ".$classpath" com.cloud.upgrade.DatabaseCreator db.properties database-imports/create-schema.sql database-imports/create-schema-premium.sql database-imports/templates.sql database-imports/developer-prefill.sql com.cloud.upgrade.DatabaseUpgradeChecker --database=cloud,usage,awsapi --rootpassword="$MYSQL_ROOT_PASSWORD"
java -cp ".$classpath" com.cloud.upgrade.DatabaseCreator db.properties database-imports/create-schema-simulator.sql database-imports/templates.simulator.sql database-imports/hypervisor_capabilities.simulator.sql com.cloud.upgrade.DatabaseUpgradeChecker --database=simulator --rootpassword="$MYSQL_ROOT_PASSWORD"
java -Xmx512m -XX:MaxPermSize=512m -jar "$jetty_runner_path" --path /client 'cloudstack.war' --lib 'lib/'
