/**
 * Re-export from refactored module for backward compatibility
 *
 * The ImportRepositoryModal has been split into:
 * - ImportRepositoryModal/ImportRepositoryModal.tsx - Main component
 * - ImportRepositoryModal/useImportWizard.ts - State management hook
 * - ImportRepositoryModal/steps/*.tsx - Individual step components
 * - ImportRepositoryModal/types.ts - TypeScript types
 */
export { ImportRepositoryModal, default } from "./ImportRepositoryModal/index";
export type { ImportRepositoryModalProps } from "./ImportRepositoryModal/index";
