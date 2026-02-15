import { SlideDeck } from '@/types/presentation';

export const exportToPPT = async (deck: SlideDeck) => {
    try {
        // Force browser-only execution
        if (typeof window === 'undefined') return;

        // Use require to bypass static analysis for node modules in some bundlers
        // But for Next.js/Webpack, dynamic import is best.
        // We need to ensure pptxgenjs is treated as a browser module.
        const pptxgen = (await import('pptxgenjs')).default;

        const pres = new pptxgen();

        // Set Presentation Title
        pres.title = deck.title;
        pres.subject = deck.description;
        pres.layout = 'LAYOUT_16x9';

        // Title Slide
        const titleSlide = pres.addSlide();
        titleSlide.addText(deck.title, { x: 1, y: 1, w: 8, h: 1, fontSize: 36, align: 'center', bold: true });
        titleSlide.addText(deck.description, { x: 1, y: 2.5, w: 8, h: 1, fontSize: 18, align: 'center' });

        // Content Slides
        deck.slides.forEach((slide) => {
            const pptSlide = pres.addSlide();

            // Add Slide Title
            pptSlide.addText(slide.title, { x: 0.5, y: 0.5, w: 9, h: 0.8, fontSize: 24, bold: true, color: '363636' });

            // Add Content based on layout
            if (slide.layout === 'bullet_points' || slide.layout === 'title_and_body') {
                if (slide.bullet_points) {
                    const bulletItems = slide.bullet_points.map((point) => ({ text: point, options: { fontSize: 18, bullet: true, breakLine: true } }));
                    pptSlide.addText(bulletItems, { x: 1, y: 1.5, w: 8, h: 4 });
                }
            } else if (slide.layout === 'chart_focus') {
                pptSlide.addText("(Chart Placeholder - Exporting charts requires image capture)", { x: 1, y: 2, w: 8, h: 1, fontSize: 14, italic: true, align: 'center' });
            }

            // Add Speaker Notes
            if (slide.speaker_notes) {
                pptSlide.addNotes(slide.speaker_notes);
            }
        });

        // Save the Presentation
        await pres.writeFile({ fileName: `${deck.title.replace(/\s+/g, '_')}.pptx` });
    } catch (error) {
        console.error("Failed to export PPT:", error);
        throw error;
    }
};
