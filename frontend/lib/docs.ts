import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

const docsDirectory = path.join(process.cwd(), 'app/docs/content');

export function getDocBySlug(slug: string) {
    const realSlug = slug.replace(/\.md$/, '');
    const fullPath = path.join(docsDirectory, `${realSlug}.md`);

    try {
        const fileContents = fs.readFileSync(fullPath, 'utf8');
        const { data, content } = matter(fileContents);

        return {
            slug: realSlug,
            frontmatter: data,
            content,
        };
    } catch (_e) {
        return null;
    }
}

export function getAllDocs() {
    const slugs = fs.readdirSync(docsDirectory);
    const docs = slugs.map((slug) => getDocBySlug(slug));
    return docs;
}
