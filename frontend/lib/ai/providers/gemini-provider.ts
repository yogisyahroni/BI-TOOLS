import { GoogleGenerativeAI } from "@google/generative-ai";
import { type AIRequest, type AIResponse, type IAIProvider } from "./base-provider";

export class GeminiProvider implements IAIProvider {
  private genAI: GoogleGenerativeAI;

  constructor(apiKey: string) {
    this.genAI = new GoogleGenerativeAI(apiKey);
  }

  async generateQuery(request: AIRequest): Promise<AIResponse> {
    const model = this.genAI.getGenerativeModel({ model: request.model.id });

    const systemPrompt = `
            Anda adalah "Expert Data & Business Analyst" tingkat senior.
            Tugas Anda adalah menerima instruksi bahasa manusia dan mengubahnya menjadi SQL query yang valid untuk database ${request.databaseType}.

            GUNAKAN SKEMA DDL BERIKUT:
            ${request.schemaDDL || "Tidak ada skema yang disediakan."}

            ATURAN:
            1. Selalu sertakan LIMIT untuk keamanan (maks 100 jika tidak diminta).
            2. Gunakan sintaks yang spesifik untuk ${request.databaseType}.
            3. Respon HARUS selalu dalam format JSON valid.

            STRUKTUR JSON:
            {
                "sql": "SELECT ...",
                "explanation": "Penjelasan singkat...",
                "confidence": 0.95,
                "suggestedVisualization": "bar" | "line" | "pie" | "table" | "metric",
                "insights": ["insight 1", "insight 2"]
            }
        `;

    const result = await model.generateContent([systemPrompt, request.prompt]);
    const responseText = result.response.text();

    try {
      // Clean markdown blocks if present
      const cleanJson = responseText.replace(/```json\n?|\n?```/g, "").trim();
      return JSON.parse(cleanJson) as AIResponse;
    } catch (e: unknown) {
      // eslint-disable-next-line no-console
      console.error("[GeminiProvider] Failed to parse AI response:", responseText, e);
      throw new Error("AI returned invalid JSON format.");
    }
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-explicit-any
  async generateChartConfig(_query: string, _data: any[]): Promise<any> {
    // This is a placeholder for future implementation
    return {};
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async generateSQL(schema: any, question: string): Promise<string> {
    // This is a placeholder for future implementation
    return `SELECT * FROM your_table WHERE 1=1 -- Based on schema: ${JSON.stringify(schema)} and question: ${question}`;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-explicit-any
  async analyzeResults(_data: any[], _query: string): Promise<string[]> {
    // Implementation for analyzing raw data results
    // This is a placeholder for future implementation
    return ["Insight analysis not yet implemented for live data."];
  }
}
