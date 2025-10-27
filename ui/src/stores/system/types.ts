import { type SystemEnvironment } from "@/domain/system";

export interface SystemEnvironmentState {
  environment: SystemEnvironment | null;
  loadingEnvironment: boolean;
  loadedEnvironment: boolean;
}

export interface SystemEnvironmentActions {
  fetchEnvironment: (refresh?: boolean) => Promise<SystemEnvironment | null>;
}

export interface SystemEnvironmentStore extends SystemEnvironmentState, SystemEnvironmentActions {}

