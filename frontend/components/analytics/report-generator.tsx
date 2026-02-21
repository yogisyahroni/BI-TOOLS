"use client";

import { useState } from "react";
import { Download, FileSpreadsheet } from "lucide-react";
import * as XLSX from "xlsx";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/use-toast";

interface ReportGeneratorProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any[];
  headers?: string[];
  title?: string;
  filename?: string;
  className?: string;
}

export function ReportGenerator({
  data,
  headers,
  title = "Dashboard Export",
  filename = "report",
  className,
}: ReportGeneratorProps) {
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleExport = async () => {
    if (!data || data.length === 0) {
      toast({
        title: "No Data",
        description: "There is no data to export.",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);

    try {
      // Determine headers if not provided
      const reportHeaders = headers || (data.length > 0 ? Object.keys(data[0]) : []);

      // Create worksheet
      const ws = XLSX.utils.json_to_sheet(data, {
        header: reportHeaders,
      });

      // Create workbook
      const wb = XLSX.utils.book_new();
      XLSX.utils.book_append_sheet(wb, ws, "Data");

      // Generate Excel file
      const excelBuffer = XLSX.write(wb, { bookType: "xlsx", type: "array" });
      const blob = new Blob([excelBuffer], {
        type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      });

      // Download file
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${filename}_${new Date().toISOString().split("T")[0]}.xlsx`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast({
        title: "Export Successful",
        description: `Downloaded ${data.length} rows to Excel.`,
      });
    } catch (error) {
      console.error(error);
      toast({
        title: "Export Failed",
        description: "Failed to generate the Excel report.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button
      variant="outline"
      size="sm"
      onClick={handleExport}
      disabled={loading || !data || data.length === 0}
      className={className}
    >
      {loading ? (
        <>Generating...</>
      ) : (
        <>
          <FileSpreadsheet className="mr-2 h-4 w-4" />
          Export to Excel
        </>
      )}
    </Button>
  );
}
