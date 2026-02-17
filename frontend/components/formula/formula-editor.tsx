"use client"

import { useState } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { Loader2, CheckCircle2, AlertCircle, Calculator, Play } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { fetchWithAuth } from "@/lib/utils"

interface ValidationResponse {
    valid: boolean
    error?: string
    references?: string[]
}

interface EvaluationResponse {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    result: any
    error?: string
}

export function FormulaEditor() {
    const [formula, setFormula] = useState("")
    const [isValidating, setIsValidating] = useState(false)
    const [isEvaluating, setIsEvaluating] = useState(false)
    const [validationResult, setValidationResult] = useState<ValidationResponse | null>(null)
    const [evaluationResult, setEvaluationResult] = useState<EvaluationResponse | null>(null)

    const handleValidate = async () => {
        setIsValidating(true)
        setValidationResult(null)
        setEvaluationResult(null)

        try {
            const res = await fetchWithAuth("/api/formulas/validate", {
                method: "POST",
                body: JSON.stringify({ formula }),
            })

            if (!res.ok) throw new Error("Validation request failed")

            const data = await res.json()
            setValidationResult(data)
        } catch (_error) {
            setValidationResult({ valid: false, error: "Network error or server unavailable" })
        } finally {
            setIsValidating(false)
        }
    }

    const handleEvaluate = async () => {
        setIsEvaluating(true)
        setEvaluationResult(null)
        // Also validate implicitly? No, let's just evaluate.

        try {
            const res = await fetchWithAuth("/api/formulas/evaluate", {
                method: "POST",
                body: JSON.stringify({ formula }),
            })

            if (!res.ok) throw new Error("Evaluation request failed")

            const data = await res.json()
            setEvaluationResult(data)
        } catch (_error) {
            setEvaluationResult({ result: null, error: "Network error or server unavailable" })
        } finally {
            setIsEvaluating(false)
        }
    }

    return (
        <div className="w-full max-w-2xl mx-auto p-4 space-y-6">
            <Card className="border-border/40 bg-card/50 backdrop-blur-sm shadow-sm">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Calculator className="w-5 h-5 text-primary" />
                        Formula Editor
                    </CardTitle>
                    <CardDescription>
                        Enter a formula to validate syntax or evaluate results.
                        Supported functions: SUM, AVG, MAX, MIN, IF, etc.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="relative">
                        <Textarea
                            placeholder="e.g., SUM(10, 20) * 2"
                            value={formula}
                            onChange={(e) => setFormula(e.target.value)}
                            className="font-mono text-base min-h-[120px] resize-y bg-background/50"
                        />
                    </div>

                    <div className="flex gap-3">
                        <Button
                            onClick={handleValidate}
                            disabled={!formula || isValidating || isEvaluating}
                            variant="secondary"
                            className="w-full sm:w-auto"
                        >
                            {isValidating ? (
                                <>
                                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                    Validating...
                                </>
                            ) : (
                                <>
                                    <CheckCircle2 className="w-4 h-4 mr-2" />
                                    Validate Syntax
                                </>
                            )}
                        </Button>

                        <Button
                            onClick={handleEvaluate}
                            disabled={!formula || isValidating || isEvaluating}
                            className="w-full sm:w-auto"
                        >
                            {isEvaluating ? (
                                <>
                                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                    Evaluating...
                                </>
                            ) : (
                                <>
                                    <Play className="w-4 h-4 mr-2" />
                                    Run Evaluation
                                </>
                            )}
                        </Button>
                    </div>
                </CardContent>
                <CardFooter className="flex-col items-stretch gap-4 border-t bg-muted/20 p-6">
                    <AnimatePresence mode="wait">
                        {validationResult && (
                            <motion.div
                                initial={{ opacity: 0, y: 10 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, y: -10 }}
                                className="w-full"
                            >
                                {validationResult.valid ? (
                                    <Alert variant="default" className="border-green-500/20 bg-green-500/10 text-green-600 dark:text-green-400">
                                        <CheckCircle2 className="h-4 w-4" />
                                        <AlertTitle>Formula is Valid</AlertTitle>
                                        <AlertDescription>
                                            References found: {validationResult.references && validationResult.references.length > 0 ? validationResult.references.join(", ") : "None"}
                                        </AlertDescription>
                                    </Alert>
                                ) : (
                                    <Alert variant="destructive">
                                        <AlertCircle className="h-4 w-4" />
                                        <AlertTitle>Syntax Error</AlertTitle>
                                        <AlertDescription>{validationResult.error}</AlertDescription>
                                    </Alert>
                                )}
                            </motion.div>
                        )}

                        {evaluationResult && (
                            <motion.div
                                initial={{ opacity: 0, y: 10 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, y: -10 }}
                                className="w-full"
                            >
                                {evaluationResult.error ? (
                                    <Alert variant="destructive">
                                        <AlertCircle className="h-4 w-4" />
                                        <AlertTitle>Evaluation Error</AlertTitle>
                                        <AlertDescription>{evaluationResult.error}</AlertDescription>
                                    </Alert>
                                ) : (
                                    <Card className="border-primary/20 bg-primary/5">
                                        <CardHeader className="py-3">
                                            <CardTitle className="text-sm font-medium text-primary">Result</CardTitle>
                                        </CardHeader>
                                        <CardContent className="py-3">
                                            <div className="text-2xl font-bold font-mono">
                                                {JSON.stringify(evaluationResult.result)}
                                            </div>
                                        </CardContent>
                                    </Card>
                                )}
                            </motion.div>
                        )}
                    </AnimatePresence>
                </CardFooter>
            </Card>
        </div>
    )
}
