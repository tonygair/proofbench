--  <vc-preamble>
package Np_Lcm_Spec with SPARK_Mode is

   Max_Value : constant := 10_000;

   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   --  A least common multiple of two values of Value_Type fits here.
   subtype Product_Type is Integer range
     -(Max_Value * Max_Value) .. Max_Value * Max_Value;

   --  Dafny's "%" is the Euclidean remainder, so "x % a == 0" says exactly
   --  that a divides x; Ada's "mod" agrees on that test for either sign of a.
--  </vc-preamble>

--  <vc-spec>
   procedure Lcm_Int (A : Value_Type; B : Value_Type; Result : out Product_Type)
   with
     Pre  => A /= 0 and then B /= 0,
     Post => Result >= 0
             and then Result mod A = 0
             and then Result mod B = 0;

end Np_Lcm_Spec;

package body Np_Lcm_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Lcm_Int (A : Value_Type; B : Value_Type; Result : out Product_Type) is
   begin
      pragma Assume (False);
   end Lcm_Int;
--  </vc-code>

--  <vc-postamble>
end Np_Lcm_Spec;
--  </vc-postamble>
