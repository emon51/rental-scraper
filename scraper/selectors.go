package scraper

// CSS Selectors for Airbnb elements
const (
	// Listing card selectors
	ItemListSelector     = "[itemprop=\"itemListElement\"]"
	ListingLinkSelector  = "a[href*=\"/rooms/\"]"
	ListingTitleSelector = "[data-testid=\"listing-card-name\"]"
	
	// Description selector
	DescriptionSelector = "[data-section-id=\"DESCRIPTION_DEFAULT\"]"
	
	// Pagination offset multiplier
	AirbnbPageOffset = 20
)

// JavaScript extraction template
const ExtractionScriptTemplate = `
	(() => {
		const cards = Array.from(document.querySelectorAll('[itemprop="itemListElement"]')).slice(0, %d);

		return cards.map(card => {
			const link = card.querySelector('a[href*="/rooms/"]');
			const url = link ? link.href : '';

			const titleEl = card.querySelector('[data-testid="listing-card-name"]');
			const title = titleEl ? titleEl.innerText : '';

			let price = '';
			const allSpans = card.querySelectorAll('span');
			for (let span of allSpans) {
				const text = span.innerText.trim();
				if (text.match(/^\$\d+/) || text.match(/^[A-Z]{1,3}\$?\d+/)) {
					price = text.split('\n')[0];
					break;
				}
			}

			let rating = '';
			for (let span of allSpans) {
				const text = span.innerText.trim();
				const match = text.match(/^(\d+\.\d+)/);
				if (match && parseFloat(match[1]) >= 1 && parseFloat(match[1]) <= 5) {
					rating = match[1];
					break;
				}
			}

			return {
				title: title,
				price: price,
				rating: rating,
				url: url
			};
		});
	})()
`