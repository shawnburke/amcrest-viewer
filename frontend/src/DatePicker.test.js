import { Time, minute } from './time'
import {datePickerController} from './DatePicker';


describe('DatePicker', () => {

    

        var dp = new datePickerController(new Time("2020-06-02"), new Time("2020-06-05"), new Time("2020-06-03"), 420);


        it('Should properly init',  () => {
            var minIso = dp.min().iso();
            var maxIso = dp.max().iso();
            var curIso = dp.date().iso();

            expect(minIso).toEqual("2020-06-02T07:00:00.000Z");
            expect(maxIso).toEqual("2020-06-05T07:00:00.000Z");
            expect(curIso).toEqual("2020-06-03T07:00:00.000Z");
        })


        it('Should properly box',  () => {

            var dp = new datePickerController(new Time("2020-06-02"), new Time("2020-06-05"), new Time("2020-06-03"), 420);

            var d = dp.date(new Time("2020-05-01"))
            expect(d.iso()).toEqual("2020-06-02T07:00:00.000Z");
           
            d = dp.date(new Time("2020-07-01"));
            expect(d.iso()).toEqual("2020-06-05T07:00:00.000Z");
        })

        it('Should properly increment and stop',  () => {

            var dp = new datePickerController(new Time("2020-06-02"), new Time("2020-06-05"), new Time("2020-06-03"), 420);

            var d = dp.date(new Time("2020-06-03"))
            expect(d.iso()).toEqual("2020-06-03T07:00:00.000Z");
           
            expect(dp.canDecrement()).toEqual(true);
            expect(dp.canIncrement()).toEqual(true);

            d = dp.advance(-1);

            expect(d.iso()).toEqual("2020-06-02T07:00:00.000Z");

            expect(dp.canDecrement()).toEqual(false);

            d = dp.advance(-1);
           
            expect(d.iso()).toEqual("2020-06-02T07:00:00.000Z");

            d = dp.advance(3);

            expect(d.iso()).toEqual("2020-06-05T07:00:00.000Z");

            expect(dp.canIncrement()).toEqual(false);
            expect(dp.canDecrement()).toEqual(true);


        })
    
});