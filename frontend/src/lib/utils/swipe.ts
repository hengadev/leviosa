import type { Attachment } from 'svelte/attachments';

// A horizontal swipe on a touch screen, for the month calendars: `{@attach
// swipe({ onLeft: next, onRight: previous })}`. Only a clearly horizontal
// stroke counts, so scrolling the page up or down over the element is left
// to the browser (give the element `touch-pan-y`); a swipe moves too far for
// the browser to treat it as a tap on what is under the finger.
export function swipe(handlers: { onLeft: () => void; onRight: () => void }): Attachment<HTMLElement> {
	return (element) => {
		let start: { x: number; y: number } | null = null;

		function onStart(event: TouchEvent) {
			start = event.touches.length === 1 ? { x: event.touches[0].clientX, y: event.touches[0].clientY } : null;
		}
		function onEnd(event: TouchEvent) {
			if (!start) return;
			const touch = event.changedTouches[0];
			const dx = touch.clientX - start.x;
			const dy = touch.clientY - start.y;
			start = null;
			if (Math.abs(dx) < 50 || Math.abs(dx) < Math.abs(dy) * 1.5) return;
			if (dx < 0) handlers.onLeft();
			else handlers.onRight();
		}
		function onCancel() {
			start = null;
		}

		element.addEventListener('touchstart', onStart, { passive: true });
		element.addEventListener('touchend', onEnd);
		element.addEventListener('touchcancel', onCancel);
		return () => {
			element.removeEventListener('touchstart', onStart);
			element.removeEventListener('touchend', onEnd);
			element.removeEventListener('touchcancel', onCancel);
		};
	};
}
