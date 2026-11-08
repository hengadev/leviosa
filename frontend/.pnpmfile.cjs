// bits-ui 2 depends on runed 0.35, whose optional @sveltejs/kit peer (only for
// runed's useSearchParams, which bits-ui never imports) would otherwise
// resolve to our dev dependency and pull SvelteKit, Vite and their
// dependencies into the production install the runtime image copies. Drop
// that peer: runed works without it.
module.exports = {
	hooks: {
		readPackage(pkg) {
			if (pkg.name === 'runed') {
				if (pkg.peerDependencies) delete pkg.peerDependencies['@sveltejs/kit'];
				if (pkg.peerDependenciesMeta) delete pkg.peerDependenciesMeta['@sveltejs/kit'];
			}
			return pkg;
		}
	}
};
