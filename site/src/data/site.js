module.exports = {
	env: process.env.NODE_ENV,
	date: Date.now(),
	githubUrl: 'https://github.com/umputun/updater',
	githubBranch: 'master',
	url: 'https://updater.umputun.dev',
	title: 'Updater',
	subtitle: 'Web-hook based receiver executing updates via HTTP request',
	description:
		'Updater is a simple web-hook-based receiver executing things via HTTP requests and invoking remote updates without exposing any sensitive info, like ssh keys, passwords, etc. The updater is usually called from CI/CD system (i.e., Github action), and the actual http call looks like curl https://<server>/update/<task-name>/<access-key>',
}
