import requests

cookie = "nyt-a=UuQD0rnfPbKryS94k3NCjJ; nyt-purr=pnhhprhosckrsdrh; nyt-jkidd=uid=86493738&lastRequest=1694066604719&activeDays=%5B1%2C0%2C0%2C0%2C1%2C0%2C0%2C1%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C0%2C1%5D&adv=4&a7dv=1&a14dv=1&a21dv=1&lastKnownType=sub&newsStartDate=1673456387&entitlements=AAA+ATH+AUD+CKG+MM+MOW+MSD+MTD+WC+XWD; datadome=3vW0WXQavqyy7xHXUT~n5kcfcWtpvFsNgEmU_4~yL49jy_XIvB7DKdcXz_0HF~4jhoGnyJ1WlB8P4~Erg_1YIad7OlD_tQG~KjmW9I3bb92D3RjKa0jNtmp3AoAo4gDK; purr-cache=<K0<rUS+US-WI<Co<G_<S0<a1<ur; nyt-auth-method=sso; nyt-m=2B89D1B1F475F6BDB1E50F91FE791B21&n=i.2&pr=l.4.0.0.0.0&imu=i.1&ira=i.0&iir=i.0&uuid=s.bdfa89cf-fccb-4ceb-b86a-788edd525dd7&igf=i.0&fv=i.0&ica=i.0&iue=i.1&iub=i.0&iga=i.0&iru=i.1&e=i.1690898400&g=i.0&ifv=i.0&v=i.0&vr=l.4.0.0.0.0&cav=i.1&igu=i.1&igd=i.0&vp=i.0&ier=i.0&ird=i.0&t=i.7&rc=i.0&ft=i.0&prt=i.0&s=s.wirecutter&er=i.1689144550&imv=i.0; datadome=7XLfZo4IN50mZpHiMjcy9BKIB_rG8Yj~aOxqB-wyOoLayTgKEnPp3MPT_UdG~705SkJlN7FdK6UipM4l3XiKGWtMxAhl32L2R9GMF7442VFra3osZeBt3FxeaO1PDDOt; NYT-S=2EplSORYSQY1EE196mg97FHFhPNMyoUao5a4Lp7zPLMwThxnwNdX1CS6OcALkxH5WhjOoea6bgYnTBH3pVcqmapsbJzoeHj.8IqKY5rjCmVUzuFykQ7CaKQtC.uZshnps6ik8zOz7RwFKv4q9f2SMirGP3pQmWLkSoIKfMxropab7d4q1iYRFi/nZ4qgtmFniukkEl16MlSfTBd741zZHjCM6s8uAZ2OHc^^^^CBUSLwiultaiBhCf0OWnBhoSMS2vFXv6XOLlfE-4XRQvYbWOIKqUnykqAh53ONf03NkFGkAVIHFzCQYP0w4JNNXID4KO3oN40ZqUFqzRMb80WqkuKeCmk9Xy0-jrH1LRHd1NG3aL9r8rXB6CZNOMpVsgNqoF; SIDNY=CBUSLwiultaiBhCf0OWnBhoSMS2vFXv6XOLlfE-4XRQvYbWOIKqUnykqAh53ONf03NkFGkAVIHFzCQYP0w4JNNXID4KO3oN40ZqUFqzRMb80WqkuKeCmk9Xy0-jrH1LRHd1NG3aL9r8rXB6CZNOMpVsgNqoF; nyt-b3-traceid=3f73bc812aac4d5198ecec75133f0e29; nyt-xwd-hashd=false; nyt-gdpr=0; nyt-geo=US; b2b_cig_opt=%7B%22isCorpUser%22%3Afalse%7D; edu_cig_opt=%7B%22isEduUser%22%3Afalse%7D"
r = requests.get(
    "https://www.nytimes.com/svc/crosswords/v6/leaderboard/mini/2023-07-02.json",
    headers={"cookie": cookie},
)

print(r.json())