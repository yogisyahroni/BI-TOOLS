import { apiGet, apiPost } from './config';
import { type DataClassification, type ColumnMetadata, type ColumnPermission } from '../../types/data-governance';

export const dataGovernanceApi = {
    getClassifications: () =>
        apiGet<DataClassification[]>('/api/governance/classifications'),

    getColumnMetadata: (datasourceID: string, tableName?: string) => {
        const url = tableName
            ? `/api/governance/metadata?datasource_id=${datasourceID}&table_name=${tableName}`
            : `/api/governance/metadata?datasource_id=${datasourceID}`;
        return apiGet<ColumnMetadata[]>(url);
    },

    updateColumnMetadata: (metadata: Partial<ColumnMetadata>) =>
        apiPost<{ success: boolean; data: ColumnMetadata }>('/api/governance/metadata', metadata),

    getColumnPermissions: (roleID: number) =>
        apiGet<ColumnPermission[]>(`/api/governance/permissions?role_id=${roleID}`),

    setColumnPermission: (permission: Partial<ColumnPermission>) =>
        apiPost<{ success: boolean }>('/api/governance/permissions', permission),
};
