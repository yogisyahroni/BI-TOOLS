export const dynamic = 'force-dynamic';

import { redirect } from 'next/navigation';

export default function QueryPage() {
    redirect('/query-builder');
}
