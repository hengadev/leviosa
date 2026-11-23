import * as env from '$app/env/private';
import { error, isRedirect, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

// --- API response types (snake_case from backend) ---

interface APISummary {
	gross_revenue_cents: number;
	refunds_cents: number;
	disputed_cents: number;
	on_site_cents: number;
	net_revenue_cents: number;
}

interface APITransaction {
	id: string;
	slot_start_time: string;
	/** The day the payment counts on: revenue is counted on the day it was paid. */
	paid_at: string;
	client_name: string;
	partner_name: string;
	product_name: string;
	amount_cents: number;
	payment_mode: 'online' | 'on_site';
	payment_status: 'paid' | 'refunded' | 'disputed';
	booking_status: 'completed' | 'cancelled' | 'confirmed' | 'no_show';
}

interface APIFinancialSummary {
	summary: APISummary;
	transactions: APITransaction[];
}

interface DashboardStatsAPI {
	bookings_this_week: number;
	revenue_this_week: number;
	upcoming_bookings_count: number;
	pending_bookings_count: number;
}

export interface Last7DaysStats {
	revenueThisWeek: number;
	bookingsThisWeek: number;
	upcomingBookingsCount: number;
	pendingBookingsCount: number;
}

// --- UI types (camelCase) ---

export interface SummaryKPIs {
	grossRevenueInCents: number;
	refundsInCents: number;
	disputedInCents: number;
	onSiteInCents: number;
	netRevenueInCents: number;
}

export interface Transaction {
	id: string;
	/** Unique row key: a booking paid online and on site has two rows. */
	key: string;
	date: string;
	description: string;
	clientName: string;
	partnerName: string;
	amountInCents: number;
	type: 'payment' | 'refund';
	paymentMode: 'online' | 'on_site';
	paymentStatus: 'paid' | 'refunded' | 'disputed';
	bookingStatus: string;
}

export interface ComptaData {
	summary: SummaryKPIs;
	transactions: Transaction[];
	from: string;
	to: string;
	last7Days: Last7DaysStats | null;
}

function getMockLast7Days(): Last7DaysStats {
	return {
		revenueThisWeek: 82400,
		bookingsThisWeek: 12,
		upcomingBookingsCount: 5,
		pendingBookingsCount: 2
	};
}

async function getMockComptaData(from: string, to: string): Promise<ComptaData> {
	const now = new Date();

	const summary: SummaryKPIs = {
		grossRevenueInCents: 62300,
		refundsInCents: 2150,
		disputedInCents: 5000,
		onSiteInCents: 15000,
		netRevenueInCents: 60150
	};

	const transactions: Transaction[] = [
		{ id: 'TXN-001', key: 'TXN-001-online', date: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString(), description: 'Massage Relaxant 60min', clientName: 'Marie Dupont', partnerName: 'Dr. Martin', amountInCents: 5000, type: 'payment', paymentMode: 'online', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-001', key: 'TXN-001-on-site', date: new Date(now.getTime() - 1 * 60 * 60 * 1000).toISOString(), description: 'Massage Relaxant 60min', clientName: 'Marie Dupont', partnerName: 'Dr. Martin', amountInCents: 7000, type: 'payment', paymentMode: 'on_site', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-002', key: 'TXN-002-online', date: new Date(now.getTime() - 4 * 60 * 60 * 1000).toISOString(), description: 'Soin du Dos 60min', clientName: 'Jean Durand', partnerName: 'Dr. Martin', amountInCents: 6500, type: 'payment', paymentMode: 'online', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-003', key: 'TXN-003-online', date: new Date(now.getTime() - 6 * 60 * 60 * 1000).toISOString(), description: 'Remboursement', clientName: 'Claire Bernard', partnerName: 'Dr. Dupont', amountInCents: 5000, type: 'refund', paymentMode: 'online', paymentStatus: 'refunded', bookingStatus: 'cancelled' },
		{ id: 'TXN-004', key: 'TXN-004-online', date: new Date(now.getTime() - 24 * 60 * 60 * 1000).toISOString(), description: 'Drainage Lymphatique', clientName: 'Lucas Petit', partnerName: 'Dr. Dupont', amountInCents: 7000, type: 'payment', paymentMode: 'online', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-005', key: 'TXN-005-online', date: new Date(now.getTime() - 48 * 60 * 60 * 1000).toISOString(), description: 'Consultation Kiné 45min', clientName: 'Thomas Richard', partnerName: 'Dr. Martin', amountInCents: 3500, type: 'payment', paymentMode: 'online', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-006', key: 'TXN-006-online', date: new Date(now.getTime() - 72 * 60 * 60 * 1000).toISOString(), description: 'Massage Relaxant 60min', clientName: 'Camille Simon', partnerName: 'Dr. Dupont', amountInCents: 5000, type: 'payment', paymentMode: 'online', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-007', key: 'TXN-007-online', date: new Date(now.getTime() - 96 * 60 * 60 * 1000).toISOString(), description: 'Remboursement', clientName: 'Louis Laurent', partnerName: 'Dr. Martin', amountInCents: 6500, type: 'refund', paymentMode: 'online', paymentStatus: 'refunded', bookingStatus: 'cancelled' },
		{ id: 'TXN-008', key: 'TXN-008-online', date: new Date(now.getTime() - 120 * 60 * 60 * 1000).toISOString(), description: 'Massage Relaxant 90min', clientName: 'Lola Bernard', partnerName: 'Dr. Dupont', amountInCents: 7000, type: 'payment', paymentMode: 'online', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-008', key: 'TXN-008-on-site', date: new Date(now.getTime() - 119 * 60 * 60 * 1000).toISOString(), description: 'Massage Relaxant 90min', clientName: 'Lola Bernard', partnerName: 'Dr. Dupont', amountInCents: 8000, type: 'payment', paymentMode: 'on_site', paymentStatus: 'paid', bookingStatus: 'completed' },
		{ id: 'TXN-009', key: 'TXN-009-online', date: new Date(now.getTime() - 144 * 60 * 60 * 1000).toISOString(), description: 'Massage Relaxant 60min', clientName: 'Nora Girard', partnerName: 'Dr. Martin', amountInCents: 5000, type: 'payment', paymentMode: 'online', paymentStatus: 'disputed', bookingStatus: 'completed' },
	];

	return { summary, transactions, from, to, last7Days: getMockLast7Days() };
}

export const load: PageServerLoad = async ({ fetch, url }) => {
	const from = url.searchParams.get('from') || defaultFrom();
	const to = url.searchParams.get('to') || defaultTo();

	if (env.USE_MOCK_DATA === 'true') {
		return await getMockComptaData(from, to);
	}

	try {
		const [summaryRes, statsRes] = await Promise.all([
			fetch(`${env.API_URL}/admin/bookings/financial-summary?from=${from}&to=${to}`),
			fetch(`${env.API_URL}/admin/dashboard/stats`).catch(() => new Response(null, { status: 503 }))
		]);
		if (summaryRes.status === 401 || summaryRes.status === 403) {
			throw redirect(302, '/auth');
		}
		if (!summaryRes.ok) {
			throw new Error(`Failed to fetch financial summary: ${summaryRes.status} ${summaryRes.statusText}`);
		}

		const api: APIFinancialSummary = await summaryRes.json();

		// The last 7 days are a bonus: the month's figures still show without them
		let last7Days: Last7DaysStats | null = null;
		try {
			if (statsRes.ok) {
				const statsApi: DashboardStatsAPI = await statsRes.json();
				last7Days = {
					revenueThisWeek: statsApi.revenue_this_week,
					bookingsThisWeek: statsApi.bookings_this_week,
					upcomingBookingsCount: statsApi.upcoming_bookings_count,
					pendingBookingsCount: statsApi.pending_bookings_count
				};
			} else {
				console.error('Error loading dashboard stats:', statsRes.status);
			}
		} catch (err) {
			console.error('Error loading dashboard stats:', err);
		}

		const summary: SummaryKPIs = {
			grossRevenueInCents: api.summary.gross_revenue_cents,
			refundsInCents: api.summary.refunds_cents,
			disputedInCents: api.summary.disputed_cents,
			onSiteInCents: api.summary.on_site_cents,
			netRevenueInCents: api.summary.net_revenue_cents
		};

		const transactions: Transaction[] = (api.transactions ?? []).map((t) => ({
			id: t.id,
			key: `${t.id}-${t.payment_mode}`,
			date: t.paid_at || t.slot_start_time,
			description: t.product_name,
			clientName: t.client_name,
			partnerName: t.partner_name,
			amountInCents: t.amount_cents,
			type: t.payment_status === 'refunded' ? 'refund' as const : 'payment' as const,
			paymentMode: t.payment_mode,
			paymentStatus: t.payment_status,
			bookingStatus: t.booking_status
		}));

		return { summary, transactions, from, to, last7Days };
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading financial summary:', err);
		throw error(503, 'Impossible de charger les données comptables. Veuillez réessayer.');
	}
};

function defaultFrom(): string {
	const now = new Date();
	return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`;
}

function defaultTo(): string {
	const now = new Date();
	return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
}
