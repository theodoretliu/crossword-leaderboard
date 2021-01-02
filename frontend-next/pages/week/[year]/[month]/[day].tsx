import { useRouter } from "next/router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { GetServerSideProps } from "next";
import { Header } from "components/header";

dayjs.extend(utc);

export const getServerSideProps: GetServerSideProps = async (context) => {
  console.log(context.query);
  return { props: {} };
};

export default function Week(props) {
  const router = useRouter();

  const { year, month, day } = router.query as {
    year: string;
    month: string;
    day: string;
  };

  const date = dayjs.utc(
    Date.UTC(parseInt(year), parseInt(month), parseInt(day))
  );

  console.log(date.toString());

  return (
    <div>
      <Header />
      Week of {date.format("dddd, MMMM D, YYYY")}
    </div>
  );
}
