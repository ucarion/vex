# vex

Vex is a feature-flagging system designed to let you change rollout strategies
without changing your code.

## Motivation

With Vex, it's possible to implement a rollout strategy with the following
phases (you can jump between these phases at will):

1. Don't deploy to any customers (i.e. "inert deploy")
2. Deploy to one "test" customer
3. Deploy to a set of "test" customers
4. Deploy to all "dogfood" customers, where "dogfood" is a standard
   set of testing customers used by your company.
5. Deploy to all dogfood customers + N% of free-tier customers
6. Deploy to all dogfood + free-tier customers + N% of premium-tier customers
7. Deploy to all dogfood + free-tier customers + N% of premium-tier customers,
   *except* for a particular set of customers
8. Deploy to everyone but a particular set of customers
9. Deploy to everyone

Rollouts like this are very common in SaaS, where you want to progressively roll
something out from low-risk to high-risk customers, or when particular customers
have usage patterns that you aren't ready to handle yet. But to support all of
these features you need:

1. The ability to change a flag from a "whitelist" to a "blacklist",
2. The ability to have a flag be "probabilistic".
3. The ability to integrate a flag with a set of data sources, such as billing
   tier. Whereas whitelist/blacklists are for a relatively fixed set
   special-case customers, the list of customers on premium tier is huge and
   ever-changing.

## Approach

Vex enables complex rollouts by separating the two concerns in feature-flagging:

* The **data plane** is the place in your code that checks whether to use a
  feature or not. Vex keeps the data plane as simple as possible. You ask Vex
  whether a flag is enabled for a customer or not. All you get is a true/false,
  or an error if the flag backend can't be reached.

  ```go
  import "github.com/ucarion/vex"

  // vex.On(...) returns (bool, error)
  showFeatureX, err := vex.On(ctx, "my-service", "my-feature", customerID)
  ```

* The **control plane** is where you *define* your flags, perhaps through a web
  UI, a CLI tool, some Slack bot, an API, or all of the above. Vex lets you
  define flags with a basic expression syntax that's just powerful enough to
  cover most use-cases.

  See the "Demo" section for an example of how you can use Vex to implement all
  9 examples at the top of this README.

* Critcally, the control plane enables **flag re-use via references**. You can
  define a flag to be a function of another flag. In particular, you can say
  things like:

  > This flag is enabled for `$customerID` if the flag in the namespace
  > `is-premium-tier` with the name `$customerID` is enabled.

  The idea being that you can ETL billing information, or whatever other
  external data sources are relevant to rollouts at your organization, into a
  namespace like `is-premium-tier`. And then teams all across your organization
  can compose their flags on top of these trusted, approved data sources.

## Demo

This repo comes with an example CLI tool that lets you create flags. To use that
tool, first run:

```bash
# This repo uses DynamoDB as a Vex store, just for demo purposes. Vex is not
# married to any particular backend.
docker-compose up -d
make ddb-create-table
```

Then, you can define a flag like so:

```bash
AWS_ACCESS_KEY_ID="x" AWS_SECRET_ACCESS_KEY="x" \
  go run ./cmd/vex/... flags create my-service my-feature ...
```

Where `...` is the flag's definition. Here's what you would put for `...` for
each of the examples in the "Motivation" section above:

1. Don't deploy to any customers (i.e. "inert deploy")

   ```json
   { "type": "constant", "constant": false }
   ```

2. Deploy to one "test" customer

   ```json
   { "type": "value_in", "value_in": ["test-customer1"] }
   ```

3. Deploy to a set of "test" customers

   ```json
   { "type": "value_in", "value_in": ["test-customer1", "test-customer2"] }
   ```

4. Deploy to all "dogfood" customers, where "dogfood" is a standard
   set of testing customers used by your company.

   ```json
   { "type": "ref", "ref": "is-dogfood" }
   ```

5. Deploy to all dogfood customers + N% of free-tier customers

   ```json
   {
     "type": "any_of",
     "any_of": [
       { "type": "ref", "ref": "is-dogfood" },
       {
         "type": "all_of",
         "all_of": [
           { "type": "ref", "ref": "is-free-tier" },
           { "type": "percent", "percent": 0.05 }
         ]
       }
     ]
   }
   ```

6. Deploy to all dogfood + free-tier customers + N% of premium-tier customers

   ```json
   {
     "type": "any_of",
     "any_of": [
       { "type": "ref", "ref": "is-dogfood" },
       { "type": "ref", "ref": "is-free-tier" },
       {
         "type": "all_of",
         "all_of": [
           { "type": "ref", "ref": "is-premium-tier" },
           { "type": "percent", "percent": 0.05 }
         ]
       }
     ]
   }
   ```

7. Deploy to all dogfood + free-tier customers + N% of premium-tier customers,
   *except* for a particular set of customers

   ```json
   {
     "type": "any_of",
     "any_of": [
       { "type": "ref", "ref": "is-dogfood" },
       { "type": "ref", "ref": "is-free-tier" },
       {
         "type": "all_of",
         "all_of": [
           { "type": "ref", "ref": "is-premium-tier" },
           { "type": "percent", "percent": 0.80 },
           {
             "type": "not",
             "not": {
               "type": "value_in", "value_in": ["exclude-1", "exclude-2"]
             }
           }
         ]
       }
     ]
   }
   ```

8. Deploy to everyone but a particular set of customers

   ```json
   {
     "type": "not",
     "not": {
       "type": "value_in", "value_in": ["exclude-1", "exclude-2"]
     }
   }
   ```

9. Deploy to everyone

   ```json
   { "type": "constant", "constant": true }
   ```

These examples assume that you've already ETL'd a set of `is-dogfood`,
`is-free-tier`, and `is-premium-tier` flags into Vex.
