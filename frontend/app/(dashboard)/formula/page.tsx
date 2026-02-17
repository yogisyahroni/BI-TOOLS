import { FormulaEditor } from "@/components/formula/formula-editor"
import { type Metadata } from "next"

export const metadata: Metadata = {
    title: "Formula Editor | InsightEngine",
    description: "Test and validate formulas",
}

export default function FormulaPage() {
    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between space-y-2">
                <h2 className="text-3xl font-bold tracking-tight">Formula Editor</h2>
            </div>
            <div className="hidden h-full flex-1 flex-col space-y-8 md:flex">
                <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-1">
                    <FormulaEditor />
                </div>
            </div>
        </div>
    )
}
