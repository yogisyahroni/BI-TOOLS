import OpenAI from "openai";
import { type AIRequest, type AIResponse, type IAIProvider } from "./base-provider";

export class OpenAIProvider implements IAIProvider {
  private client: OpenAI;

  constructor(apiKey: string) {
    this.client = new OpenAI({ apiKey });
  }

  async generateQuery(request: AIRequest): Promise<AIResponse> {
    const systemPrompt = `
            Anda adalah "Expert Data & Business Analyst" tingkat senior.
            Tugas Anda adalah menghasilkan query ${request.databaseType} berdasarkan skema DDL yang diberikan.
            
            DDL SCHEMA:
            ${request.schemaDDL}

            OUTPUT FORMAT: Valid JSON only.
        `;

    const completion = await this.client.chat.completions.create({
      model: request.model.id,
      messages: [
        { role: "system", content: systemPrompt },
        { role: "user", content: request.prompt },
      ],
      response_format: { type: "json_object" },
    });

    const content = completion.choices[0].message.content;
    if (!content) throw new Error("OpenAI returned empty response.");

    return JSON.parse(content) as AIResponse;
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async analyzeResults(_data: any[], _sql: string): Promise<string[]> {
    return ["Insight analysis pending."];
  }

  async explainSQL(_sql: string): Promise<string> {
    return "Insight analysis pending.";
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unused-vars
  async generateSQL(_schema: any, _question: string): Promise<string> {
    return "SQL generation pending.";
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unused-vars
  async generateChartConfig(_query: string, _data: any[]): Promise<any> {
    return {};
  }
}
