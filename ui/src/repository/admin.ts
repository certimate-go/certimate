import { COLLECTION_NAME_ADMIN, getPocketBase } from "./_pocketbase";

const pb = getPocketBase();
const pbco = pb.collection(COLLECTION_NAME_ADMIN);

export const authWithPassword = (username: string, password: string) => {
  return pbco.authWithPassword(username, password);
};

export const getAuthStore = () => {
  return pb.authStore;
};

export const save = (data: { email: string } | { password: string; passwordConfirm: string }) => {
  return pbco.update(getAuthStore().record?.id || "", data);
};
