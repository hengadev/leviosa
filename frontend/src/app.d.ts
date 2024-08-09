// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user: {
				email: string;
				lastname: string;
				firstname: string;
				gender: string;
				birthdate: string;
				telephone: string;
				address: string;
				city: string;
				postalcard: string;
			};
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
