type AccordionContent = {
	trigger: string;
	content: string;
};

const accordionItems: AccordionContent[] = [
	{
		trigger: 'Comment reserver ?',
		content: 'Yes. It adheres to the WAI-ARIA design pattern.'
	},
	{
		trigger: 'Vote pour plannifier tes evenements !',
		content:
			'Chaque mois, un certain nombre de dates sont disponibles au vote des utilisateurs.Seulement celles ayant recu le plus de vote seront retenus.'
	},
	{
		trigger: 'Comment reserver ?',
		content: "Choisis une date ainsi qu'un creneau qui t'interesse et fais ta reservation."
	},
	{
		trigger: 'De quoi ai-je besoin ?',
		content: "Une tenue de sport des plus simples, de quoi s'hydrater et un esprit positif"
	}
];

export function load({ data }) {
	// the data comes from the +page.server.ts
	const { nextVotes } = data;
	return { accordionItems, nextVotes };
}
