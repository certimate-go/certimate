import { getPocketBase } from "./_pocketbase";

const pb = getPocketBase();

export const listCronJobs = () => {
  return pb.crons
    .getFullList({
      requestKey: null,
    })
    .then((res) => {
      const jobs = res
        .filter((job) => !job.id.startsWith("__pb"))
        .map((job) => {
          return {
            id: job.id,
            cron: job.expression,
          };
        });
      return {
        items: jobs,
      };
    });
};

export type ListLogsRequest = {
  page?: number;
  perPage?: number;
};

export const listLogs = (request: ListLogsRequest) => {
  const page = request.page || 1;
  const perPage = request.perPage || 10;

  return pb.logs
    .getList(page, perPage, {
      filter: 'data.type!="request"',
      sort: "-@rowid",
      skipTotal: true,
      requestKey: null,
    })
    .then((res) => {
      return {
        items: res.items,
      };
    });
};
